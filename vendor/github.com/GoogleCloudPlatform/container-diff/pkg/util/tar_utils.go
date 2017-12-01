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

package util

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func unpackTar(tr *tar.Reader, path string) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			logrus.Error("Error getting next tar header")
			return err
		}

		if strings.Contains(header.Name, ".wh.") {
			rmPath := filepath.Join(path, header.Name)
			newName := strings.Replace(rmPath, ".wh.", "", 1)
			if err := os.Remove(rmPath); err != nil {
				logrus.Error(err)
			}
			if err = os.RemoveAll(newName); err != nil {
				logrus.Error(err)
			}
			continue
		}

		target := filepath.Join(path, header.Name)
		mode := header.FileInfo().Mode()
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); os.IsNotExist(err) {
				if err := os.MkdirAll(target, mode); err != nil {
					return err
				}
			} else {
				if err := os.Chmod(target, mode); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			// It's possible for a file to be included before the directory it's in is created.
			baseDir := filepath.Dir(target)
			if _, err := os.Stat(baseDir); os.IsNotExist(err) {
				if err := os.MkdirAll(baseDir, 0755); err != nil {
					return err
				}
			}
			currFile, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
			if err != nil {
				logrus.Errorf("Error opening file %s", target)
				return err
			}
			// manually set permissions on file, since the default umask (022) will interfere
			if err = os.Chmod(target, mode); err != nil {
				logrus.Errorf("Error updating file permissions on %s", target)
				return err
			}
			_, err = io.Copy(currFile, tr)
			if err != nil {
				return err
			}
			currFile.Close()
		}

	}
	return nil
}

// UnTar takes in a path to a tar file and writes the untarred version to the provided target.
// Only untars one level, does not untar nested tars.
func UnTar(r io.Reader, target string) error {
	if _, ok := os.Stat(target); ok != nil {
		os.MkdirAll(target, 0775)
	}

	tr := tar.NewReader(r)
	if err := unpackTar(tr, target); err != nil {
		return err
	}
	return nil
}

func IsTar(path string) bool {
	return filepath.Ext(path) == ".tar" ||
		filepath.Ext(path) == ".tar.gz" ||
		filepath.Ext(path) == ".tgz"
}

func CheckTar(image string) bool {
	if strings.TrimSuffix(image, ".tar") == image {
		return false
	}
	if _, err := os.Stat(image); err != nil {
		logrus.Errorf("%s does not exist", image)
		return false
	}
	return true
}
