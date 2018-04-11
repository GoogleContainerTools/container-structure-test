/*
Copyright 2017 Google, Inc. All rights reserved.

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

package image

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/containers/image/manifest"
	"github.com/containers/image/types"
	digest "github.com/opencontainers/go-digest"
)

type fields struct {
	mfst *manifest.Schema2
	cfg  *manifest.Schema2Image
}
type args struct {
	content string
}

var testCases = []struct {
	name    string
	fields  fields
	args    args
	wantErr bool
}{
	{
		name: "add layer",
		fields: fields{
			mfst: &manifest.Schema2{
				LayersDescriptors: []manifest.Schema2Descriptor{
					{
						Digest: digest.Digest("abc123"),
					},
				},
			},
			cfg: &manifest.Schema2Image{
				RootFS: &manifest.Schema2RootFS{
					DiffIDs: []digest.Digest{digest.Digest("bcd234")},
				},
				History: []manifest.Schema2History{
					{
						CreatedBy: "foo",
					},
				},
			},
		},
		args: args{
			content: "myextralayer",
		},
		wantErr: false,
	},
}

func TestMutableSource_AppendLayer(t *testing.T) {
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			m := &MutableSource{
				mfst:       tt.fields.mfst,
				cfg:        tt.fields.cfg,
				extraBlobs: make(map[string][]byte),
			}

			if err := m.AppendLayer([]byte(tt.args.content), "container-diff"); (err != nil) != tt.wantErr {
				t.Fatalf("MutableSource.AppendLayer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := m.saveConfig(); err != nil {
				t.Fatalf("Error saving config: %v", err)
			}
			// One blob for the new layer, one for the new config.
			if len(m.extraBlobs) != 2 {
				t.Fatal("No extra blob stored after appending layer.")
			}

			r, _, err := m.GetBlob(types.BlobInfo{Digest: m.mfst.ConfigDescriptor.Digest})
			if err != nil {
				t.Fatal("Not able to get new config blob.")
			}

			cfgBytes, err := ioutil.ReadAll(r)
			if err != nil {
				t.Fatal("Unable to read config.")
			}
			cfg := manifest.Schema2Image{}
			if err := json.Unmarshal(cfgBytes, &cfg); err != nil {
				t.Fatal("Unable to parse config.")
			}

			if len(cfg.History) != 2 {
				t.Fatalf("No layer added to image history: %v", cfg.History)
			}

			if len(cfg.RootFS.DiffIDs) != 2 {
				t.Fatalf("No layer added to Diff IDs: %v", cfg.RootFS.DiffIDs)
			}
			if cfg.RootFS.DiffIDs[1] != digest.FromString(tt.args.content) {
				t.Fatalf("Incorrect diffid for content. Expected %s, got %s", digest.FromString(tt.args.content), cfg.RootFS.DiffIDs[1])
			}
		})
	}
}

func TestMutableSource_Env(t *testing.T) {

	cfg := &manifest.Schema2Image{
		Schema2V1Image: manifest.Schema2V1Image{
			Config: &manifest.Schema2Config{
				Env: []string{
					"PATH=/path/to/dir",
				},
			},
		},
	}

	m := &MutableSource{
		mfst:       &manifest.Schema2{},
		cfg:        cfg,
		extraBlobs: make(map[string][]byte),
	}
	initialEnvMap := m.Env()
	expectedInitialEnvMap := map[string]string{
		"PATH": "/path/to/dir",
	}
	if !reflect.DeepEqual(initialEnvMap, expectedInitialEnvMap) {
		t.Fatalf("Got incorrect environment map, got: %s, expected: %s", initialEnvMap, expectedInitialEnvMap)
	}

	initialEnvMap["NEW"] = "new"

	m.SetEnv(initialEnvMap, "container-diff")

	newEnvMap := m.Env()
	expectedNewEnvMap := map[string]string{
		"PATH": "/path/to/dir",
		"NEW":  "new",
	}
	if !reflect.DeepEqual(newEnvMap, expectedNewEnvMap) {
		t.Fatalf("Got incorrect environment map, got: %s, expected: %s", newEnvMap, expectedNewEnvMap)
	}

	// Ensure length of history is 1
	if len(cfg.History) != 1 {
		t.Fatalf("No layer added to image history: %v", cfg.History)
	}
}

func TestMutableSource_Config(t *testing.T) {
	cfg := &manifest.Schema2Image{
		Schema2V1Image: manifest.Schema2V1Image{
			Config: &manifest.Schema2Config{},
		},
	}

	m := &MutableSource{
		mfst:       &manifest.Schema2{},
		cfg:        cfg,
		extraBlobs: make(map[string][]byte),
	}
	config := m.Config()
	user := "new-user"
	config.User = user
	m.SetConfig(config, "container-diff", true)
	// Ensure length of history is 1
	if len(cfg.History) != 1 {
		t.Fatalf("No layer added to image history: %v", cfg.History)
	}
	// Ensure command was set
	if !reflect.DeepEqual(m.Config().User, user) {
		t.Fatalf("Config command incorrectly set, was: %s, expected: %s", m.Config().User, user)
	}

}
