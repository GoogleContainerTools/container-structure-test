package render

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/GoogleCloudPlatform/runtimes-common/docgen/lib/proto"
)

const DOCKER = "DOCKER"
const DOCKER_COMPOSE = "DOCKER_COMPOSE"
const KUBERNETES = "KUBERNETES"

type Runtime proto.Runtime

func (t Runtime) IsDocker() bool {
	return proto.Runtime(t) == proto.Runtime_DOCKER
}

func (t Runtime) IsKubernetes() bool {
	return proto.Runtime(t) == proto.Runtime_KUBERNETES
}

func (t Runtime) String() string {
	return proto.Runtime(t).String()
}

func (t Runtime) AnchorIdSuffix() string {
	switch proto.Runtime(t) {
	case proto.Runtime_DOCKER:
		return "docker"
	case proto.Runtime_KUBERNETES:
		return "kubernetes"
	default:
		panic(fmt.Sprintf("Unhandled Runtime: %v", t))
	}
}

type Document struct {
	*proto.Document
	AnchorIds map[string]bool
}

func NewDocument(document *proto.Document) *Document {
	anchorIds := make(map[string]bool)
	for _, taskGroup := range document.TaskGroups {
		anchorIds[taskGroup.AnchorId] = true
		for _, task := range taskGroup.Tasks {
			anchorIds[task.AnchorId] = true
		}
	}

	result := &Document{document, anchorIds}
	return result
}

func (t *Document) NeedsGcloud() bool {
	gcloudRegex := regexp.MustCompile(`\bgcloud\b`)
	return gcloudRegex.FindString(t.Overview.PullCommand) != ""
}

func (t *Document) HasReferences() bool {
	return t.EnvironmentVariableReference != nil || t.PortReference != nil || t.VolumeReference != nil
}

func (t *Document) PortReference() *PortReference {
	if t.Document.PortReference == nil {
		return nil
	}
	return &PortReference{t.Document.PortReference}
}

func (t *Document) DockerTaskGroups() []*SingleRuntimeTaskGroup {
	return t.taskGroupsFor(proto.Runtime_DOCKER)
}

func (t *Document) KubernetesTaskGroups() []*SingleRuntimeTaskGroup {
	return t.taskGroupsFor(proto.Runtime_KUBERNETES)
}

func (t *Document) taskGroupsFor(runtime proto.Runtime) []*SingleRuntimeTaskGroup {
	result := make([]*SingleRuntimeTaskGroup, 0, len(t.Document.TaskGroups))
	for _, tg := range t.Document.TaskGroups {
		taskGroup := &TaskGroup{tg, t}
		srtg := taskGroup.ForRuntime(runtime)
		if srtg != nil {
			result = append(result, srtg)
		}
	}
	return result
}

func (t *Document) expandAnchors(text string, runtime Runtime) string {
	refsRegex := regexp.MustCompile(`\[\]\(#.+\)`)
	singleRefRegex := regexp.MustCompile(`\[\]\(#(.+)(?:|(.+))?\)`)
	return refsRegex.ReplaceAllStringFunc(text, func(m string) string {
		anchorId := string(singleRefRegex.ExpandString(
			[]byte{}, "$1", m, singleRefRegex.FindStringSubmatchIndex(m)))
		forcedRuntime := string(singleRefRegex.ExpandString(
			[]byte{}, "$2", m, singleRefRegex.FindStringSubmatchIndex(m)))
		if len(forcedRuntime) > 0 {
			switch forcedRuntime {
			case DOCKER:
				runtime = Runtime(proto.Runtime_DOCKER)
			case KUBERNETES:
				runtime = Runtime(proto.Runtime_KUBERNETES)
			default:
				panic(fmt.Sprintf("Unknown runtime in anchor: %s", m))
			}
		}
		return fmt.Sprintf(
			"[%s](#%s-%s)", t.getTitleForAnchor(anchorId), anchorId, strings.ToLower(runtime.String()))
	})
}

func (t *Document) getTitleForAnchor(anchorId string) string {
	for _, taskGroup := range t.TaskGroups {
		if taskGroup.AnchorId == anchorId {
			return taskGroup.Title
		}
		for _, task := range taskGroup.Tasks {
			if task.AnchorId == anchorId {
				return task.Title
			}
		}
	}
	panic(fmt.Sprintf("Unable to find task or task group for anchor: %s", anchorId))
}

type TaskGroup struct {
	*proto.TaskGroup
	Document *Document
}

func (tg *TaskGroup) ForRuntime(runtime proto.Runtime) *SingleRuntimeTaskGroup {
	tasks := make([]*proto.Task, 0, len(tg.TaskGroup.Tasks))
	for _, t := range tg.TaskGroup.Tasks {
		task := &Task{t, tg.Document}
		if task.HasRuntime(runtime) {
			tasks = append(tasks, task.ForRuntime(runtime).Task)
		}
	}
	if len(tasks) > 0 {
		srtg := &SingleRuntimeTaskGroup{&proto.TaskGroup{}, Runtime(runtime), tg.Document}
		*srtg.TaskGroup = *tg.TaskGroup
		srtg.TaskGroup.Tasks = tasks
		return srtg
	}
	return nil
}

// SingleRuntimeTaskGroup wraps a TaskGroup proto whose tasks are
// only applicable to one specific Runtime.
type SingleRuntimeTaskGroup struct {
	*proto.TaskGroup
	Runtime  Runtime
	Document *Document
}

func (tg *SingleRuntimeTaskGroup) Tasks() []*SingleRuntimeTask {
	result := make([]*SingleRuntimeTask, 0, len(tg.TaskGroup.Tasks))
	for _, t := range tg.TaskGroup.Tasks {
		result = append(result, &SingleRuntimeTask{t, tg.Runtime, tg.Document})
	}
	return result
}

func (t *SingleRuntimeTaskGroup) AnchorId() string {
	return fmt.Sprintf("%s-%v", t.TaskGroup.AnchorId, t.Runtime.AnchorIdSuffix())
}

type Task struct {
	*proto.Task
	Document *Document
}

func (t *Task) HasRuntime(runtime proto.Runtime) bool {
	for _, rt := range t.Runtimes {
		if rt == runtime {
			return true
		}
	}
	return false
}

func (t *Task) ForRuntime(runtime proto.Runtime) *SingleRuntimeTask {
	return NewSingleRuntimeTask(t.Task, Runtime(runtime), t.Document)
}

// SingleRuntimeTask wraps a Task proto whose instructions are
// applicable to only one specific Runtime.
type SingleRuntimeTask struct {
	*proto.Task
	Runtime  Runtime
	Document *Document
}

func NewSingleRuntimeTask(task *proto.Task, runtime Runtime, document *Document) *SingleRuntimeTask {
	result := &SingleRuntimeTask{&proto.Task{}, Runtime(runtime), document}
	*result.Task = *task
	result.Task.Instructions = make([]*proto.TaskInstruction, 0, len(task.Instructions))
	for _, ins := range task.Instructions {
		isApplicable := len(ins.ApplicableRuntimes) == 0
		if len(ins.ApplicableRuntimes) > 0 {
			for _, rt := range ins.ApplicableRuntimes {
				if rt == proto.Runtime(runtime) {
					isApplicable = true
					break
				}
			}
		}
		if isApplicable {
			result.Task.Instructions = append(result.Task.Instructions, ins)
		}
	}
	return result
}

func (t *SingleRuntimeTask) Instructions() []*TaskInstruction {
	result := make([]*TaskInstruction, 0, len(t.Task.Instructions))
	for _, i := range t.Task.Instructions {
		result = append(result, &TaskInstruction{i, t.Runtime, t.Document})
	}
	return result
}

func (t *SingleRuntimeTask) AnchorId() string {
	return fmt.Sprintf("%s-%v", t.Task.AnchorId, t.Runtime.AnchorIdSuffix())
}

type TaskInstruction struct {
	*proto.TaskInstruction
	Runtime  Runtime
	Document *Document
}

func (t *TaskInstruction) GetRun() *RunInstruction {
	if t.TaskInstruction.GetRun() == nil {
		return nil
	}
	return &RunInstruction{t.TaskInstruction.GetRun(), t.Runtime, t}
}

func (t *TaskInstruction) GetExec() *ExecInstruction {
	if t.TaskInstruction.GetExec() == nil {
		return nil
	}
	return &ExecInstruction{t.TaskInstruction.GetExec(), t.Runtime, t}
}

func (t *TaskInstruction) GetDockerfile() *DockerfileInstruction {
	if t.TaskInstruction.GetDockerfile() == nil {
		return nil
	}
	return &DockerfileInstruction{t.TaskInstruction.GetDockerfile(), t}
}

func (t *TaskInstruction) GetCopy() *CopyInstruction {
	if t.TaskInstruction.GetCopy() == nil {
		return nil
	}
	return &CopyInstruction{t.TaskInstruction.GetCopy(), t.Runtime, t}
}

func (t *TaskInstruction) Description() string {
	return t.Document.expandAnchors(t.TaskInstruction.Description, t.Runtime)
}

type RunInstruction struct {
	*proto.RunInstruction
	Runtime         Runtime
	TaskInstruction *TaskInstruction
}

func (t *RunInstruction) RunType() RunInstruction_RunType {
	return RunInstruction_RunType(t.RunInstruction.RunType)
}

// ContainerName constructs the name for a stand-alone
// docker container or k8s pod.
func (t *RunInstruction) ContainerName() string {
	return "some-" + t.Name
}

func (t *RunInstruction) Dependencies() []*RunInstruction_Dependency {
	result := make([]*RunInstruction_Dependency, 0, len(t.RunInstruction.Dependencies))
	for _, dependency := range t.RunInstruction.Dependencies {
		result = append(result, &RunInstruction_Dependency{dependency, t.Runtime})
	}
	return result
}

func (t *RunInstruction) DependenciesWithLinkAlias() []*RunInstruction_Dependency {
	result := make([]*RunInstruction_Dependency, 0, len(t.Dependencies()))
	for _, dependency := range t.Dependencies() {
		if len(dependency.DockerLinkAlias) > 0 && dependency.DockerLinkAlias != dependency.Name {
			result = append(result, dependency)
		}
	}
	return result
}

func (t *RunInstruction) DependenciesWithoutLinkAlias() []*RunInstruction_Dependency {
	result := make([]*RunInstruction_Dependency, 0, len(t.Dependencies()))
	for _, dependency := range t.Dependencies() {
		if len(dependency.DockerLinkAlias) == 0 || dependency.DockerLinkAlias == dependency.Name {
			result = append(result, dependency)
		}
	}
	return result
}

func (t *RunInstruction) DockerEnvironment() map[string]string {
	return makeEnvironmentVariablesMap(t.RunInstruction.Environment, DOCKER)
}

func (t *RunInstruction) DockerComposeEnvironment() map[string]string {
	return makeEnvironmentVariablesMap(t.RunInstruction.Environment, DOCKER_COMPOSE)
}

func (t *RunInstruction) KubernetesEnvironment() map[string]string {
	return makeEnvironmentVariablesMap(t.RunInstruction.Environment, KUBERNETES)
}

func (t *RunInstruction) ExposedPorts() []*RunInstruction_ExposedPort {
	result := make([]*RunInstruction_ExposedPort, 0, len(t.RunInstruction.ExposedPorts))
	for _, exposedPort := range t.RunInstruction.ExposedPorts {
		result = append(result, &RunInstruction_ExposedPort{exposedPort})
	}
	return result
}

func (t *RunInstruction) MappedExposedPorts() []*RunInstruction_ExposedPort {
	result := make([]*RunInstruction_ExposedPort, 0, len(t.ExposedPorts()))
	for _, exposedPort := range t.ExposedPorts() {
		port := *exposedPort
		if exposedPort.Mapped == 0 {
			port.Mapped = port.Port
		}
		result = append(result, &port)
	}
	return result
}

func (t *RunInstruction) ConcatArguments() string {
	// TODO: Quote arguments if needed.
	return strings.Join(t.Arguments, " ")
}

func (t *RunInstruction) Volumes() []*RunInstruction_MountedVolume {
	result := make([]*RunInstruction_MountedVolume, 0, len(t.RunInstruction.Volumes))
	for _, volume := range t.RunInstruction.Volumes {
		result = append(result, &RunInstruction_MountedVolume{volume})
	}
	return result
}

// AllVolumes returns volumes across the main service and all dependencies.
func (t *RunInstruction) AllVolumes() []*RunInstruction_MountedVolume {
	result := make([]*RunInstruction_MountedVolume, 0, len(t.RunInstruction.Volumes))
	for _, volume := range t.RunInstruction.Volumes {
		result = append(result, &RunInstruction_MountedVolume{volume})
	}
	for _, dependency := range t.Dependencies() {
		for _, volume := range dependency.Volumes() {
			result = append(result, volume)
		}
	}
	return result
}

// AllEmptyPersistentVolumes returns empty persistent volumes across the
// main service and all dependencies.
func (t *RunInstruction) AllEmptyPersistentVolumes() []*RunInstruction_MountedVolume {
	result := make([]*RunInstruction_MountedVolume, 0, len(t.Volumes()))
	for _, volume := range t.Volumes() {
		if volume.GetEmptyPersistentVolume() != nil {
			result = append(result, volume)
		}
	}
	for _, dependency := range t.Dependencies() {
		for _, volume := range dependency.Volumes() {
			if volume.GetEmptyPersistentVolume() != nil {
				result = append(result, volume)
			}
		}
	}
	return result
}

// SingleFileVolumes returns single file volumes across the main service
// and all dependencies.
func (t *RunInstruction) AllSingleFileVolumes() []*RunInstruction_MountedVolume {
	result := make([]*RunInstruction_MountedVolume, 0, len(t.Volumes()))
	for _, volume := range t.Volumes() {
		if volume.GetSingleFile() != nil {
			result = append(result, volume)
		}
	}
	for _, dependency := range t.Dependencies() {
		for _, volume := range dependency.Volumes() {
			if volume.GetSingleFile() != nil {
				result = append(result, volume)
			}
		}
	}
	return result
}

type RunInstruction_Dependency struct {
	*proto.RunInstruction_Dependency
	Runtime Runtime
}

func (t *RunInstruction_Dependency) ContainerName() string {
	return "some-" + t.Name
}

func (t *RunInstruction_Dependency) Volumes() []*RunInstruction_MountedVolume {
	result := make([]*RunInstruction_MountedVolume, 0, len(t.RunInstruction_Dependency.Volumes))
	for _, volume := range t.RunInstruction_Dependency.Volumes {
		result = append(result, &RunInstruction_MountedVolume{volume})
	}
	return result
}

func (t *RunInstruction_Dependency) DockerEnvironment() map[string]string {
	return makeEnvironmentVariablesMap(t.Environment, DOCKER)
}

func (t *RunInstruction_Dependency) DockerComposeEnvironment() map[string]string {
	return makeEnvironmentVariablesMap(t.Environment, DOCKER_COMPOSE)
}

func (t *RunInstruction_Dependency) KubernetesEnvironment() map[string]string {
	return makeEnvironmentVariablesMap(t.Environment, KUBERNETES)
}

type RunInstruction_RunType proto.RunInstruction_RunType

func (t RunInstruction_RunType) DetachedContainer() bool {
	return t.LongRunning()
}

func (t RunInstruction_RunType) AutoremovedContainer() bool {
	return t.Oneshot() || t.Interactive()
}

func (t RunInstruction_RunType) LongRunning() bool {
	return proto.RunInstruction_RunType(t) == proto.RunInstruction_LONG_RUNNING
}

func (t RunInstruction_RunType) Oneshot() bool {
	return proto.RunInstruction_RunType(t) == proto.RunInstruction_ONESHOT
}

func (t RunInstruction_RunType) Interactive() bool {
	return proto.RunInstruction_RunType(t) == proto.RunInstruction_INTERACTIVE_SHELL
}

type RunInstruction_ExposedPort struct {
	*proto.RunInstruction_ExposedPort
}

func (t *RunInstruction_ExposedPort) DockerPortMappingProtocol() string {
	switch t.Protocol {
	case proto.RunInstruction_ExposedPort_TCP:
		return ""
	case proto.RunInstruction_ExposedPort_UDP:
		return "/udp"
	default:
		panic(fmt.Sprintf("Unrecognized ExposedPort protocol %v", t.Protocol))
	}
}

func (t *RunInstruction_ExposedPort) NamingPort() string {
	switch t.Protocol {
	case proto.RunInstruction_ExposedPort_TCP:
		return fmt.Sprintf("%v", t.Port)
	case proto.RunInstruction_ExposedPort_UDP:
		return fmt.Sprintf("udp%v", t.Port)
	default:
		panic(fmt.Sprintf("Unrecognized ExposedPort protocol %v", t.Protocol))
	}
}

type RunInstruction_MountedVolume struct {
	*proto.RunInstruction_MountedVolume
}

func (t *RunInstruction_MountedVolume) KubernetesName() string {
	r := strings.NewReplacer(" ", "-", "_", "-")
	return strings.ToLower(r.Replace(t.Name))
}

func (t *RunInstruction_MountedVolume) KubernetesClaimName() string {
	return t.KubernetesName()
}

func (t *RunInstruction_MountedVolume) GetEmptyPersistentVolume() *RunInstruction_MountedVolume_EmptyPersistentVolume {
	p := t.RunInstruction_MountedVolume
	if p.GetEmptyPersistentVolume() == nil {
		return nil
	}
	return &RunInstruction_MountedVolume_EmptyPersistentVolume{p.GetEmptyPersistentVolume(), t}
}

func (t *RunInstruction_MountedVolume) GetSingleFile() *RunInstruction_MountedVolume_SingleFile {
	p := t.RunInstruction_MountedVolume
	if p.GetSingleFile() == nil {
		return nil
	}
	return &RunInstruction_MountedVolume_SingleFile{p.GetSingleFile(), t}
}

type RunInstruction_MountedVolume_EmptyPersistentVolume struct {
	*proto.RunInstruction_MountedVolume_EmptyPersistentVolume
	Volume *RunInstruction_MountedVolume
}

func (t *RunInstruction_MountedVolume_EmptyPersistentVolume) HostPath() string {
	p := t.RunInstruction_MountedVolume_EmptyPersistentVolume
	if len(p.HostPath) > 0 {
		return p.HostPath
	}
	return "/some/path/to/" + t.Volume.Name
}

type RunInstruction_MountedVolume_SingleFile struct {
	*proto.RunInstruction_MountedVolume_SingleFile
	Volume *RunInstruction_MountedVolume
}

func (t *RunInstruction_MountedVolume_SingleFile) HostFile() string {
	f := t.RunInstruction_MountedVolume_SingleFile.HostFile
	if strings.HasPrefix(f, "/") {
		return f
	}
	return "$(pwd)/" + f
}

func (t *RunInstruction_MountedVolume_SingleFile) HostFileBaseName() string {
	_, file := filepath.Split(t.HostFile())
	return file
}

func (t *RunInstruction_MountedVolume_SingleFile) ConfigMapName() string {
	return t.Volume.Name
}

type DockerfileInstruction struct {
	*proto.DockerfileInstruction
	TaskInstruction *TaskInstruction
}

func (t *DockerfileInstruction) DeriveTargetImage() string {
	if len(t.TargetImage) > 0 {
		return t.TargetImage
	}
	splits := strings.Split(t.BaseImage, "/")
	last := splits[len(splits)-1]
	return fmt.Sprintf("my-%v", last)
}

type ExecInstruction struct {
	*proto.ExecInstruction
	Runtime         Runtime
	TaskInstruction *TaskInstruction
}

func (t *ExecInstruction) Interactive() bool {
	return t.ExecType == proto.ExecInstruction_INTERACTIVE_SHELL
}

func (t *ExecInstruction) ContainerName() string {
	if t.GetContainerFromRun() != nil {
		runInstruction := &RunInstruction{t.GetContainerFromRun(), t.Runtime, t.TaskInstruction}
		return runInstruction.ContainerName()
	}
	return t.GetContainerName()
}

type CopyInstruction struct {
	*proto.CopyInstruction
	Runtime         Runtime
	TaskInstruction *TaskInstruction
}

func (t *CopyInstruction) ContainerName() string {
	if t.GetContainerFromRun() != nil {
		runInstruction := &RunInstruction{t.GetContainerFromRun(), t.Runtime, t.TaskInstruction}
		return runInstruction.ContainerName()
	}
	return t.GetContainerName()
}

func (t *CopyInstruction) ToContainer() bool {
	return t.Direction == proto.CopyInstruction_TO_CONTAINER
}

type PortReference struct {
	*proto.PortReference
}

func (t *PortReference) Ports() []*PortReference_PortInfo {
	result := make([]*PortReference_PortInfo, 0, len(t.PortReference.Ports))
	for _, port := range t.PortReference.Ports {
		result = append(result, &PortReference_PortInfo{port})
	}
	return result
}

type PortReference_PortInfo struct {
	*proto.PortReference_PortInfo
}

func (t *PortReference_PortInfo) UppercasedProtocol() string {
	return proto.PortReference_PortInfo_Protocol_name[int32(t.Protocol)]
}

func makeEnvironmentVariablesMap(
	env map[string]*proto.RunInstruction_EnvironmentVariableValue,
	runtime string) map[string]string {
	var valueFunc func(*proto.RunInstruction_EnvironmentVariableValue) string
	switch runtime {
	case DOCKER:
		valueFunc = func(v *proto.RunInstruction_EnvironmentVariableValue) string {
			return v.DockerValue
		}
	case DOCKER_COMPOSE:
		valueFunc = func(v *proto.RunInstruction_EnvironmentVariableValue) string {
			return v.DockerComposeValue
		}
	case KUBERNETES:
		valueFunc = func(v *proto.RunInstruction_EnvironmentVariableValue) string {
			return v.KubernetesValue
		}
	}

	result := make(map[string]string)
	for k, v := range env {
		if len(valueFunc(v)) > 0 {
			result[k] = valueFunc(v)
		} else if len(v.Value) > 0 {
			result[k] = v.Value
		}
	}
	return result
}
