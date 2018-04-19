/*
Copyright 2018 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package notify

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/config"
	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var timeLayout = time.RFC1123

func ShouldCheckURLVersion(filePath string) bool {
	if !viper.GetBool(config.UpdateCheckConfigKey) {
		return false
	}
	lastUpdateTime := getTimeFromFileIfExists(filePath)
	return time.Since(lastUpdateTime).Seconds() >= viper.GetFloat64(config.UpdateCheckIntervalInSecs)
}

type Release struct {
	Name      string
	Checksums map[string]string
}

type Releases []Release

func getJson(url string, target *Releases) error {
	r, err := http.Get(url)
	if err != nil {
		return errors.Wrap(err, "Error getting minikube version url via http")
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func GetLatestVersionFromURL(url string, versionPrefix string) (semver.Version, error) {
	r, err := getAllVersionsFromURL(url)
	if err != nil {
		return semver.Version{}, err
	}
	return semver.Make(strings.TrimPrefix(r[0].Name, versionPrefix))
}

func getAllVersionsFromURL(url string) (Releases, error) {
	var releases Releases
	if err := getJson(url, &releases); err != nil {
		return releases, errors.Wrap(err, "Error getting json from version url")
	}
	if len(releases) == 0 {
		return releases, errors.Errorf("There were no json releases at the url specified: %s", url)
	}
	return releases, nil
}

func WriteTimeToFile(path string, inputTime time.Time) error {
	err := ioutil.WriteFile(path, []byte(inputTime.Format(timeLayout)), 0644)
	if err != nil {
		return errors.Wrap(err, "Error writing current update time to file: ")
	}
	return nil
}

func getTimeFromFileIfExists(path string) time.Time {
	lastUpdateCheckTime, err := ioutil.ReadFile(path)
	if err != nil {
		return time.Time{}
	}
	timeInFile, err := time.Parse(timeLayout, string(lastUpdateCheckTime))
	if err != nil {
		return time.Time{}
	}
	return timeInFile
}
