// Copyright 2018 Google Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/GoogleContainerTools/container-structure-test/pkg/config"
	"github.com/GoogleContainerTools/container-structure-test/pkg/drivers"
	"github.com/GoogleContainerTools/container-structure-test/pkg/output"
	"github.com/GoogleContainerTools/container-structure-test/pkg/types"
	"github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func ValidateArgs(opts *config.StructureTestOptions) error {
	if opts.Driver == drivers.Host {
		if opts.Metadata == "" {
			return fmt.Errorf("Please provide path to image metadata file")
		}
		if opts.ImagePath != "" {
			return fmt.Errorf("Cannot provide both image path and metadata file")
		}
	} else {
		if opts.ImagePath == "" && opts.ImageFromLayout == "" {
			return fmt.Errorf("Please supply path to image or oci image layout to test against")
		}
		if opts.Metadata != "" {
			return fmt.Errorf("Cannot provide both image path and metadata file")
		}
	}
	if len(opts.ConfigFiles) == 0 {
		return fmt.Errorf("Please provide at least one test config file")
	}
	return nil
}

func Parse(fp string, args *drivers.DriverConfig, driverImpl func(drivers.DriverConfig) (drivers.Driver, error)) (types.StructureTest, error) {
	testContents, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	// We first have to unmarshal to determine the schema version, then we unmarshal again
	// to do the full parse.
	var unmarshal types.Unmarshaller
	var strictUnmarshal types.Unmarshaller
	var versionHolder types.SchemaVersion

	switch {
	case strings.HasSuffix(fp, ".json"):
		unmarshal = json.Unmarshal
		strictUnmarshal = json.Unmarshal
	case strings.HasSuffix(fp, ".yaml"):
		unmarshal = yaml.Unmarshal
		strictUnmarshal = yaml.UnmarshalStrict
	case strings.HasSuffix(fp, ".yml"):
		unmarshal = yaml.Unmarshal
		strictUnmarshal = yaml.UnmarshalStrict
	default:
		return nil, errors.New("Please provide valid JSON or YAML config file")
	}

	if err := unmarshal(testContents, &versionHolder); err != nil {
		return nil, err
	}

	version := versionHolder.SchemaVersion
	if version == "" {
		return nil, errors.New("Please provide schema version")
	}

	var st types.StructureTest
	if schemaVersion, ok := types.SchemaVersions[version]; ok {
		st = schemaVersion()
	} else {
		return nil, errors.New("Unsupported schema version: " + version)
	}

	if err = strictUnmarshal(testContents, st); err != nil {
		return nil, errors.New("error unmarshalling config: " + err.Error())
	}

	st.SetDriverImpl(driverImpl, *args)
	return st, nil
}

func ProcessResults(out io.Writer, format unversioned.OutputValue, c chan interface{}) error {
	totalPass := 0
	totalFail := 0
	totalDuration := time.Duration(0)
	errStrings := make([]string, 0)
	results, err := channelToSlice(c)
	if err != nil {
		return errors.Wrap(err, "reading results from channel")
	}
	for _, r := range results {
		if format == unversioned.Text {
			// output individual results if we're not in json mode
			output.OutputResult(out, r)
		}
		if r.IsPass() {
			totalPass++
		} else {
			totalFail++
		}
		totalDuration += r.Duration
	}
	if totalPass+totalFail == 0 || totalFail > 0 {
		errStrings = append(errStrings, "FAIL")
	}
	if len(errStrings) > 0 {
		err = fmt.Errorf(strings.Join(errStrings, "\n"))
	}

	summary := unversioned.SummaryObject{
		Total:    totalFail + totalPass,
		Pass:     totalPass,
		Fail:     totalFail,
		Duration: totalDuration,
	}
	if format == unversioned.Json || format == unversioned.Junit {
		// only output results here if we're in json mode
		summary.Results = results
	}
	output.FinalResults(out, format, summary)

	return err
}

func channelToSlice(c chan interface{}) ([]*unversioned.TestResult, error) {
	results := []*unversioned.TestResult{}
	for elem := range c {
		elem, ok := elem.(*unversioned.TestResult)
		if !ok {
			return nil, fmt.Errorf("unexpected value found in channel: %v", elem)
		}
		results = append(results, elem)
	}
	return results, nil
}
