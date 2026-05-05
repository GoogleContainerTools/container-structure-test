// Copyright 2026 Google Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package drivers

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/GoogleContainerTools/container-structure-test/internal/pkgutil"
	"github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	kexec "k8s.io/client-go/util/exec"
)

func findFile(file string) (string, error) {
	stat, err := os.Stat(file)
	if err == nil && !stat.IsDir() {
		return file, nil
	}

	// fallback to path relative to executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		fallbackPath := filepath.Join(exeDir, file)
		if stat, err := os.Stat(fallbackPath); err == nil && !stat.IsDir() {
			return fallbackPath, nil
		}
	}

	// fallback to user home
	if homeDir, err := os.UserHomeDir(); err == nil {
		fallbackPath := filepath.Join(homeDir, file)
		if stat, err := os.Stat(fallbackPath); err == nil && !stat.IsDir() {
			return fallbackPath, nil
		}
	}

	return "", os.ErrNotExist
}

// PodFromFile will read a given file with pod definition and returns a corev1.Pod
// accordingly.
func PodFromFile(file string) (*corev1.Pod, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	resolvedPodTemplateFile, err := findFile(file)
	if err != nil {
		return nil, err
	}

	stream, err := os.ReadFile(resolvedPodTemplateFile)
	if err != nil {
		return nil, err
	}
	obj, gvk, err := decode(stream, nil, nil)
	if err != nil {
		return nil, err
	}
	if gvk.Kind == "Pod" {
		return obj.(*corev1.Pod), nil
	}
	return nil, fmt.Errorf("invalid podtemplate: %s", file)
}

type K8sDriver struct {
	Image       string
	Client      kubernetes.Interface
	runOpts     unversioned.ContainerRunOptions
	KubeConfig  *rest.Config
	PodTemplate *corev1.Pod
	podName     string
	env         []unversioned.EnvVar
}

func NewK8sDriver(args DriverConfig) (Driver, error) {
	var k8sConfig *rest.Config
	var err error

	// Check if running inside a cluster
	if _, err = os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		k8sConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		// Not in cluster, use kubeconfig
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		if kc := os.Getenv("KUBECONFIG"); kc != "" {
			kubeconfig = kc
		}
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	template, err := PodFromFile(args.PodTemplate)
	if err != nil {
		return nil, err
	}

	return &K8sDriver{
		Client:      clientset,
		Image:       args.Image,
		runOpts:     args.RunOpts,
		PodTemplate: template,
		KubeConfig:  k8sConfig,
		podName:     "",
	}, nil
}

func (d *K8sDriver) Setup(envVars []unversioned.EnvVar, fullCommands [][]string) error {
	return nil
}

func (d *K8sDriver) Teardown(fullCommands [][]string) error {
	return nil
}

func (d *K8sDriver) SetEnv(envVars []unversioned.EnvVar) error {
	d.env = append(d.env, envVars...)
	return nil
}

func (d *K8sDriver) ProcessCommand(envVars []unversioned.EnvVar, fullCommand []string) (string, string, int, error) {
	if d.podName == "" {
		logrus.Infof("k8s driver creating pod in namespace %s", d.PodTemplate.ObjectMeta.Namespace)
		allEnvs := append(d.env, envVars...)
		// create a pod that waits
		pod, err := d.createPod(allEnvs, []string{"sleep", "99999"})
		if err != nil {
			return "", "", 0, err
		}

		// Wait for the pod to be running
		err = d.waitForPod()
		if err != nil {
			return "", "", 0, err
		}

		d.podName = pod.Name
		logrus.Infof("k8s driver created pod %s in namespace %s", d.podName, d.PodTemplate.ObjectMeta.Namespace)
	}

	return d.execInPod(fullCommand)
}

func (d *K8sDriver) StatFile(path string) (os.FileInfo, error) {
	command := []string{"stat", "-c", "%n,%s,%F,%a,%u,%g", path}
	stdout, _, _, err := d.ProcessCommand(nil, command)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(stdout), ",")
	if len(parts) != 6 {
		return nil, fmt.Errorf("unexpected output from stat: %s", stdout)
	}
	size, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}
	isDir := parts[2] == "directory"

	uid, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		return nil, err
	}

	gid, err := strconv.ParseInt(parts[5], 10, 64)
	if err != nil {
		return nil, err
	}

	// Use bitSize 32 because os.FileMode is a uint32
	fileMode64, err := strconv.ParseUint(parts[3], 8, 32)
	if err != nil {
		return nil, err
	}

	fileMode := os.FileMode(fileMode64)

	// Manually map the Linux special bits to Go constants
	if fileMode64&01000 != 0 { // Linux Sticky Bit
		fileMode |= os.ModeSticky
	}

	if isDir {
		fileMode |= os.ModeDir
	}

	return &fileInfo{
		name:     parts[0],
		size:     size,
		isDir:    isDir,
		uid:      uid,
		gid:      gid,
		fileMode: fileMode,
	}, nil
}

// fileInfo implements os.FileInfo for files in a container.
type fileInfo struct {
	name     string
	size     int64
	isDir    bool
	uid      int64
	gid      int64
	fileMode os.FileMode
}

func (fi *fileInfo) Name() string {
	return fi.name
}
func (fi *fileInfo) Size() int64 {
	return fi.size
}
func (fi *fileInfo) Mode() os.FileMode {
	return fi.fileMode
}
func (fi *fileInfo) ModTime() time.Time {
	return time.Now()
}
func (fi *fileInfo) IsDir() bool {
	return fi.isDir
}
func (fi *fileInfo) Sys() interface{} {
	hdr := &tar.Header{
		Name:    fi.name,
		Size:    fi.size,
		Mode:    int64(fi.fileMode),
		Uid:     int(fi.uid),
		Gid:     int(fi.gid),
		ModTime: fi.ModTime(),
	}
	if fi.isDir {
		hdr.Typeflag = tar.TypeDir
	} else {
		hdr.Typeflag = tar.TypeReg
	}
	return hdr
}

func (d *K8sDriver) ReadFile(path string) ([]byte, error) {
	command := []string{"cat", path}
	stdout, _, _, err := d.ProcessCommand(nil, command)
	if err != nil {
		return nil, err
	}
	return []byte(stdout), nil
}

func (d *K8sDriver) ReadDir(path string) ([]os.FileInfo, error) {
	command := []string{"ls", "-l", path}
	stdout, _, _, err := d.ProcessCommand(nil, command)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(stdout, "\n")
	var files []os.FileInfo
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 9 {
			continue
		}
		size, _ := strconv.ParseInt(parts[4], 10, 64)
		files = append(files, &fileInfo{
			name:  parts[8],
			size:  size,
			isDir: parts[0][0] == 'd',
		})
	}
	return files, nil
}

func (d *K8sDriver) GetConfig() (unversioned.Config, error) {
	// similar as tar

	imageObj, err := pkgutil.GetImageForName("remote://" + d.Image)
	if err != nil {
		return unversioned.Config{}, errors.Wrap(err, "retrieving image")
	}

	configFile, err := imageObj.Image.ConfigFile()
	if err != nil {
		return unversioned.Config{}, errors.Wrap(err, "retrieving config file")
	}
	config := configFile.Config

	// docker provides these as maps (since they can be mapped in docker run commands)
	// since this will never be the case when built through a dockerfile, we convert to list of strings
	volumes := []string{}
	for v := range config.Volumes {
		volumes = append(volumes, v)
	}

	ports := []string{}
	for p := range config.ExposedPorts {
		// docker always appends the protocol to the port, so this is safe
		ports = append(ports, strings.Split(p, "/")[0])
	}

	return unversioned.Config{
		Env:          convertSliceToMap(config.Env),
		Entrypoint:   config.Entrypoint,
		Cmd:          config.Cmd,
		Volumes:      volumes,
		Workdir:      config.WorkingDir,
		ExposedPorts: ports,
		Labels:       config.Labels,
		User:         config.User,
	}, nil
}

func (d *K8sDriver) Destroy() {
	d.stopPod()
}

func (d *K8sDriver) stopPod() {
	if d.podName == "" {
		logrus.Info("k8s driver no pod defined, nothing to clean")
		return
	}
	err := d.Client.CoreV1().Pods(d.PodTemplate.Namespace).Delete(context.Background(), d.podName, metav1.DeleteOptions{})
	if err != nil {
		logrus.Warnf("Error when removing pod %s: %s", d.podName, err.Error())
	}
	d.podName = ""
	logrus.Infof("Stopped pod %s", d.podName)
}

func (d *K8sDriver) createPod(envVars []unversioned.EnvVar, command []string) (*corev1.Pod, error) {
	// create env vars
	var env []corev1.EnvVar
	for _, envVar := range envVars {
		env = append(env, corev1.EnvVar{
			Name:  envVar.Key,
			Value: envVar.Value,
		})
	}

	// take the template, make a copy and overwrite some fields
	pod := d.PodTemplate.DeepCopy()
	pod.ObjectMeta.Name = pod.ObjectMeta.Name + "-" + rand.String(5)
	pod.Spec.Containers = []corev1.Container{
		{
			Name:            "main",
			Image:           d.Image,
			Command:         command,
			Env:             env,
			ImagePullPolicy: corev1.PullAlways,
		},
	}
	pod.Spec.RestartPolicy = corev1.RestartPolicyNever

	// store the name for later reference
	d.podName = pod.ObjectMeta.Name

	return d.Client.CoreV1().Pods(pod.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
}

func (d *K8sDriver) waitForPod() error {
	watcher, err := d.Client.CoreV1().Pods(d.PodTemplate.Namespace).Watch(context.Background(), metav1.ListOptions{
		FieldSelector: "metadata.name=" + d.podName,
	})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for event := range watcher.ResultChan() {
		pod, ok := event.Object.(*corev1.Pod)
		if !ok {
			continue
		}
		switch pod.Status.Phase {
		case corev1.PodRunning, corev1.PodSucceeded, corev1.PodFailed:
			return nil
		}
	}
	return nil
}

func (d *K8sDriver) execInPod(command []string) (string, string, int, error) {
	// TTY must be false to keep stdout and stderr separate
	req := d.Client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(d.podName).
		Namespace(d.PodTemplate.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command:   command,
			Container: "main",
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(d.KubeConfig, "POST", req.URL())
	if err != nil {
		return "", "", 0, err
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	// Handle Exit Code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(kexec.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			panic(err) // Connection or protocol error
		}
	}

	logrus.Infof("Executed command %v in %s", command, d.podName)

	return stdout.String(), stderr.String(), exitCode, nil
}
