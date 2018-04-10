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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/config"
	"github.com/blang/semver"
	"github.com/spf13/viper"
)

func TestShouldCheckURL(t *testing.T) {
	tempDir, _ := ioutil.TempDir("", "tests")
	defer os.RemoveAll(tempDir)

	lastUpdateCheckFilePath := filepath.Join(tempDir, "last_update_check")

	// test that if users disable update notification in config, the URL version does not get checked
	viper.Set(config.UpdateCheckConfigKey, false)
	if ShouldCheckURLVersion(lastUpdateCheckFilePath) {
		t.Fatalf("shouldCheckURLVersion returned true even though config had UpdateCheckConfigKey: false")
	}

	// test that if users want update notification, the URL version does get checked
	viper.Set(config.UpdateCheckConfigKey, true)
	if !ShouldCheckURLVersion(lastUpdateCheckFilePath) {
		t.Fatalf("shouldCheckURLVersion returned false even though there was no last_update_check file")
	}

	// test that update notifications get triggered if it has been longer than 24 hours
	viper.Set(config.UpdateCheckIntervalInSecs, 86400)
	WriteTimeToFile(lastUpdateCheckFilePath, time.Time{}) //time.Time{} returns time -> January 1, year 1, 00:00:00.000000000 UTC.
	if !ShouldCheckURLVersion(lastUpdateCheckFilePath) {
		t.Fatalf("shouldCheckURLVersion returned false even though longer than 24 hours since last update")
	}

	// test that update notifications do not get triggered if it has been less than 24 hours
	WriteTimeToFile(lastUpdateCheckFilePath, time.Now().UTC())
	if ShouldCheckURLVersion(lastUpdateCheckFilePath) {
		t.Fatalf("shouldCheckURLVersion returned true even though less than 24 hours since last update")
	}

}

type URLHandlerCorrect struct {
	releases Releases
}

func (h *URLHandlerCorrect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(h.releases)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprintf(w, string(b))
}

func TestGetLatestVersionFromURLCorrect(t *testing.T) {
	// test that the version is correctly parsed if returned if valid JSON is returned the url endpoint
	latestVersionFromURL := "0.0.0-dev"
	VersionPrefix := "v"
	handler := &URLHandlerCorrect{
		releases: []Release{{Name: VersionPrefix + latestVersionFromURL}},
	}
	server := httptest.NewServer(handler)

	latestVersion, err := GetLatestVersionFromURL(server.URL, VersionPrefix)
	if err != nil {
		t.Fatalf(err.Error())
	}
	expectedVersion, _ := semver.Make(latestVersionFromURL)
	if latestVersion.Compare(expectedVersion) != 0 {
		t.Fatalf("Expected latest version from URL to be %s, it was instead %s", expectedVersion, latestVersion)
	}
}

type URLHandlerNone struct{}

func (h *URLHandlerNone) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func TestGetLatestVersionFromURLNone(t *testing.T) {
	// test that an error is returned if nothing is returned at the url endpoint
	handler := &URLHandlerNone{}
	server := httptest.NewServer(handler)

	_, err := GetLatestVersionFromURL(server.URL, "")
	if err == nil {
		t.Fatalf("No version value was returned from URL but no error was thrown")
	}
}

type URLHandlerMalformed struct{}

func (h *URLHandlerMalformed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprintf(w, "Malformed JSON")
}

func TestGetLatestVersionFromURLMalformed(t *testing.T) {
	// test that an error is returned if malformed JSON is at the url endpoint
	handler := &URLHandlerMalformed{}
	server := httptest.NewServer(handler)

	_, err := GetLatestVersionFromURL(server.URL, "")
	if err == nil {
		t.Fatalf("Malformed version value was returned from URL but no error was thrown")
	}
}
