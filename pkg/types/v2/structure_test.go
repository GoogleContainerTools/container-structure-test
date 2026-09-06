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

package v2

import (
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/GoogleContainerTools/container-structure-test/pkg/drivers"
	types "github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	"github.com/GoogleContainerTools/container-structure-test/testutil"
)

// driverRecorder counts how many drivers a run asks for. For the tar driver every one of
// them unpacks the whole image into a fresh temporary directory (NewTarDriver ->
// pkgutil.GetImageForName) and Destroy deletes it again, so this count is the number of
// times the image under test is extracted.
type driverRecorder struct {
	created   int
	destroyed int
}

func (r *driverRecorder) newDriver(drivers.DriverConfig) (drivers.Driver, error) {
	r.created++
	return &recordingDriver{recorder: r}, nil
}

// recordingDriver answers every read the non-command tests make, so a run reaches its end
// and the only thing a case has to look at is how many drivers it took.
type recordingDriver struct {
	recorder *driverRecorder
}

func (d *recordingDriver) Setup(_ []types.EnvVar, _ [][]string) error { return nil }

func (d *recordingDriver) Teardown(_ [][]string) error { return nil }

func (d *recordingDriver) SetEnv(_ []types.EnvVar) error { return nil }

func (d *recordingDriver) ProcessCommand(_ []types.EnvVar, _ []string) (string, string, int, error) {
	return "", "", 0, nil
}

func (d *recordingDriver) StatFile(_ string) (os.FileInfo, error) {
	return recordedFileInfo{}, nil
}

func (d *recordingDriver) ReadFile(_ string) ([]byte, error) {
	return []byte("recorded contents"), nil
}

func (d *recordingDriver) ReadDir(_ string) ([]os.FileInfo, error) { return nil, nil }

func (d *recordingDriver) GetConfig() (types.Config, error) {
	return types.Config{
		Env:        map[string]string{"PATH": "/usr/bin"},
		Entrypoint: []string{"/entrypoint.sh"},
		Labels:     map[string]string{"maintainer": "container-structure-test"},
	}, nil
}

func (d *recordingDriver) Destroy() { d.recorder.destroyed++ }

type recordedFileInfo struct{}

func (recordedFileInfo) Name() string       { return "recorded" }
func (recordedFileInfo) Size() int64        { return 0 }
func (recordedFileInfo) Mode() fs.FileMode  { return 0644 }
func (recordedFileInfo) ModTime() time.Time { return time.Time{} }
func (recordedFileInfo) IsDir() bool        { return false }
func (recordedFileInfo) Sys() interface{}   { return nil }

func nonCommandTests(recorder *driverRecorder) *StructureTest {
	entrypoint := []string{"/entrypoint.sh"}
	st := &StructureTest{
		SchemaVersion: "2.0.0",
		FileExistenceTests: []FileExistenceTest{
			{Name: "first file", Path: "/first", ShouldExist: true, Uid: defaultOwnership, Gid: defaultOwnership},
			{Name: "second file", Path: "/second", ShouldExist: true, Uid: defaultOwnership, Gid: defaultOwnership},
			{Name: "third file", Path: "/third", ShouldExist: true, Uid: defaultOwnership, Gid: defaultOwnership},
		},
		FileContentTests: []FileContentTest{
			{Name: "first contents", Path: "/first", ExpectedContents: []string{"recorded"}},
			{Name: "second contents", Path: "/second", ExpectedContents: []string{"recorded"}},
		},
		MetadataTest: MetadataTest{
			Entrypoint: &entrypoint,
			Labels:     []types.Label{{Key: "maintainer", Value: "container-structure-test"}},
		},
	}
	st.SetDriverImpl(recorder.newDriver, drivers.DriverConfig{})
	return st
}

func drainResults(t *testing.T, channel chan interface{}) {
	t.Helper()
	close(channel)
	for result := range channel {
		if reported, ok := result.(*types.TestResult); ok && !reported.IsPass() {
			t.Errorf("%s did not pass: %v", reported.Name, reported.Errors)
		}
	}
}

// The tests that only READ an image -- file existence, file content and metadata -- share
// one driver for the whole run. One driver per test costs one full image extraction per
// test under the tar driver, so an N-test run unpacks the same image N times.
//
// Command tests are deliberately not covered here: they mutate the container they run in,
// so their per-test driver lifecycle is what isolates them from each other.
func TestReadOnlyTestsShareOneDriverForTheWholeRun(t *testing.T) {
	t.Parallel()

	recorder := &driverRecorder{}
	st := nonCommandTests(recorder)

	channel := make(chan interface{}, 32)
	st.RunAll(channel, "structure_test.go")
	drainResults(t, channel)

	testutil.CheckDeepEqual(t, 1, recorder.created)
	testutil.CheckDeepEqual(t, 1, recorder.destroyed)
}

// Each runner still stands on its own for a caller that drives one kind directly: it
// builds a driver, uses it for every test of that kind, and destroys it.
func TestARunnerDrivenOnItsOwnBuildsOneDriver(t *testing.T) {
	t.Parallel()

	recorder := &driverRecorder{}
	st := nonCommandTests(recorder)

	channel := make(chan interface{}, 32)
	st.RunFileExistenceTests(channel)
	drainResults(t, channel)

	testutil.CheckDeepEqual(t, 1, recorder.created)
	testutil.CheckDeepEqual(t, 1, recorder.destroyed)
}

// Every driver a run builds is also thrown away: leaking one would leave an extracted
// image behind on disk.
func TestEveryDriverARunBuildsIsDestroyed(t *testing.T) {
	t.Parallel()

	recorder := &driverRecorder{}
	st := nonCommandTests(recorder)

	channel := make(chan interface{}, 32)
	st.RunAll(channel, "structure_test.go")
	drainResults(t, channel)

	testutil.CheckDeepEqual(t, recorder.created, recorder.destroyed)
}
