/*
Command line tool for generating a Cloud Build yaml file based on versions.yaml.
*/
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/GoogleCloudPlatform/runtimes-common/versioning/versions"
)

type cloudBuildOptions struct {
	// Whether to restrict to a particular set of Dockerfile directories.
	// If empty, all directories are used.
	Directories []string

	// Whether to run tests as part of the build.
	RunTests bool

	// Whether to require that image tags do not already exist in the repo.
	RequireNewTags bool

	// Whether to push to all declared tags
	FirstTagOnly bool

	// Optional timeout duration. If not specified, the Cloud Builder default timeout is used.
	TimeoutSeconds int

	// Optional parallel build. If specified, images can be build on bigger machines in parallel.
	EnableParallel bool

	// Forces parallel build. If specified, images are build on bigger machines in parallel. Overrides EnableParallel.
	ForceParallel bool
}

// TODO(huyhg): Replace "gcr.io/$PROJECT_ID/functional_test" with gcp-runtimes one.
const cloudBuildTemplateString = `steps:
{{- $parallel := .Parallel }}
{{- if .RequireNewTags }}
# Check if tags exist.
{{- range .Images }}
  - name: gcr.io/gcp-runtimes/check_if_tag_exists
    args:
      - 'python'
      - '/main.py'
      - '--image={{ . }}'
{{- end }}
{{- end }}

# Build images
{{- range .ImageBuilds }}
  - name: gcr.io/cloud-builders/docker
    args:
      - 'build'
      - '--tag={{ .Tag }}'
      - '{{ .Directory }}'
{{- if $parallel }}
    waitFor: ['-']
    id: 'image-{{ .Tag }}'
{{- end }}
{{- end }}

{{- range $imageIndex, $image := .ImageBuilds }}
{{- $primary := $image.Tag }}
{{- range $testIndex, $test := $image.StructureTests }}
{{- if and (eq $imageIndex 0) (eq $testIndex 0) }}

# Run structure tests
{{- end}}
  - name: gcr.io/gcp-runtimes/structure_test
    args:
      - '--image'
      - '{{ $primary }}'
      - '--config'
      - '{{ $test }}'
{{- end }}
{{- end }}

{{- range $imageIndex, $image := .ImageBuilds }}
{{- $primary := $image.Tag }}
{{- range $testIndex, $test := $image.FunctionalTests }}
{{- if and (eq $imageIndex 0) (eq $testIndex 0) }}

# Run functional tests
{{- end }}
  - name: gcr.io/$PROJECT_ID/functional_test
    args:
      - '--verbose'
      - '--vars'
      - 'IMAGE={{ $primary }}'
      - '--vars'
      - 'UNIQUE={{ $imageIndex }}-{{ $testIndex }}'
      - '--test_spec'
      - '{{ $test }}'
{{- if $parallel }}
    waitFor: ['image-{{ $primary }}']
    id: 'test-{{ $primary }}-{{ $testIndex }}'
{{- end }}
{{- end }}

{{- end }}

# Add alias tags
{{- range $imageIndex, $image := .ImageBuilds }}
{{- $primary := $image.Tag }}
{{- range .Aliases }}
  - name: gcr.io/cloud-builders/docker
    args:
      - 'tag'
      - '{{ $primary }}'
      - '{{ . }}'
{{- if $parallel }}
    waitFor:
      - 'image-{{ $primary }}'
{{- range $testIndex, $test := $image.FunctionalTests }}
      - 'test-{{ $primary }}-{{ $testIndex }}'
{{- end }}
{{- end }}
{{- end }}
{{- end }}

images:
{{- range .AllImages }}
  - '{{ . }}'
{{- end }}

{{- if not (eq .TimeoutSeconds 0) }}

timeout: {{ .TimeoutSeconds }}s
{{- end }}

{{- if $parallel }}
options:
  machineType: 'N1_HIGHCPU_8'
{{- end }}
`

const testsDir = "tests"
const functionalTestsDir = "tests/functional_tests"
const structureTestsDir = "tests/structure_tests"
const testJsonSuffix = "_test.json"
const testYamlSuffix = "_test.yaml"
const workspacePrefix = "/workspace/"

type imageBuildTemplateData struct {
	Directory       string
	Tag             string
	Aliases         []string
	StructureTests  []string
	FunctionalTests []string
}

type cloudBuildTemplateData struct {
	RequireNewTags bool
	Parallel       bool
	ImageBuilds    []imageBuildTemplateData
	AllImages      []string
	TimeoutSeconds int
}

func shouldParallelize(options cloudBuildOptions, numberOfVersions int, numberOfTests int) bool {
	if options.ForceParallel {
		return true
	}
	if !options.EnableParallel {
		return false
	}
	return numberOfVersions > 1 || numberOfTests > 1
}

func newCloudBuildTemplateData(
	registry string, spec versions.Spec, options cloudBuildOptions) cloudBuildTemplateData {
	data := cloudBuildTemplateData{}
	data.RequireNewTags = options.RequireNewTags

	// Determine the set of directories to operate on.
	dirs := make(map[string]bool)
	if len(options.Directories) > 0 {
		for _, d := range options.Directories {
			dirs[d] = true
		}
	} else {
		for _, v := range spec.Versions {
			dirs[v.Dir] = true
		}
	}

	// Extract tests to run.
	var structureTests []string
	var functionalTests []string
	if options.RunTests {
		// Legacy structure tests reside in the root tests/ directory.
		structureTests = append(structureTests, readTests(testsDir)...)
		structureTests = append(structureTests, readTests(structureTestsDir)...)
		functionalTests = append(functionalTests, readTests(functionalTestsDir)...)
	}

	// Extract a list of full image names to build.
	for _, v := range spec.Versions {
		if !dirs[v.Dir] {
			continue
		}
		var images []string
		for _, t := range v.Tags {
			image := fmt.Sprintf("%v/%v:%v", registry, v.Repo, t)
			images = append(images, image)
			if options.FirstTagOnly {
				break
			}
		}
		data.AllImages = append(data.AllImages, images...)
		versionSTests, versionFTests := filterTests(structureTests, functionalTests, v)
		data.ImageBuilds = append(
			data.ImageBuilds, imageBuildTemplateData{v.Dir, images[0], images[1:], versionSTests, versionFTests})
	}

	data.TimeoutSeconds = options.TimeoutSeconds
	data.Parallel = shouldParallelize(options, len(spec.Versions), len(functionalTests))
	return data
}

func readTests(testsDir string) (tests []string) {
	if info, err := os.Stat(testsDir); err == nil && info.IsDir() {
		files, err := ioutil.ReadDir(testsDir)
		check(err)
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			if strings.HasSuffix(f.Name(), testJsonSuffix) || strings.HasSuffix(f.Name(), testYamlSuffix) {
				tests = append(tests, workspacePrefix+fmt.Sprintf("%s/%s", testsDir, f.Name()))
			}
		}
	}
	return
}

func filterTests(structureTests []string, functionalTests []string, version versions.Version) (outStructureTests []string, outFunctionalTests []string) {
	included := make(map[string]bool, len(structureTests)+len(functionalTests))
	for _, test := range append(structureTests, functionalTests...) {
		included[test] = true
	}
	for _, excluded := range version.ExcludeTests {
		if !included[workspacePrefix+excluded] {
			log.Fatalf("No such test to exclude: %s", excluded)
		}
		included[workspacePrefix+excluded] = false
	}

	outStructureTests = make([]string, 0, len(structureTests))
	for _, test := range structureTests {
		if included[test] {
			outStructureTests = append(outStructureTests, test)
		}
	}
	outFunctionalTests = make([]string, 0, len(functionalTests))
	for _, test := range functionalTests {
		if included[test] {
			outFunctionalTests = append(outFunctionalTests, test)
		}
	}
	return
}

func renderCloudBuildConfig(
	registry string, spec versions.Spec, options cloudBuildOptions) string {
	data := newCloudBuildTemplateData(registry, spec, options)
	tmpl, _ := template.
		New("cloudBuildTemplate").
		Parse(cloudBuildTemplateString)
	var result bytes.Buffer
	tmpl.Execute(&result, data)
	return result.String()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	registryPtr := flag.String("registry", "gcr.io/$PROJECT_ID", "Registry, e.g: 'gcr.io/my-project'")
	dirsPtr := flag.String("dirs", "", "Comma separated list of Dockerfile dirs to use.")
	testsPtr := flag.Bool("tests", true, "Run tests.")
	newTagsPtr := flag.Bool("new_tags", false, "Require that image tags do not already exist.")
	firstTagOnly := flag.Bool("first_tag", false, "Build only the first per version.")
	timeoutPtr := flag.Int("timeout", 0, "Timeout in seconds. If not set, the default Cloud Build timeout is used.")
	enableParallel := flag.Bool("enable_parallel", false, "Enable parallel build and bigger VM")
	forceParallel := flag.Bool("force_parallel", false, "Force parallel build and bigger VM")
	flag.Parse()

	if *registryPtr == "" {
		log.Fatalf("--registry flag is required")
	}

	if strings.Contains(*registryPtr, ":") {
		*registryPtr = strings.Replace(*registryPtr, ":", "/", 1)
	}

	var dirs []string
	if *dirsPtr != "" {
		dirs = strings.Split(*dirsPtr, ",")
	}
	spec := versions.LoadVersions("versions.yaml")
	options := cloudBuildOptions{dirs, *testsPtr, *newTagsPtr, *firstTagOnly, *timeoutPtr, *enableParallel, *forceParallel}
	result := renderCloudBuildConfig(*registryPtr, spec, options)
	fmt.Println(result)
}
