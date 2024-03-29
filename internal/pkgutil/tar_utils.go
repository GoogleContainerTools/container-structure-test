/*
Copyright 2018 Google, Inc. All rights reserved.

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

package pkgutil

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type OriginalPerm struct {
	path string
	perm os.FileMode
}

func unpackTar(tr *tar.Reader, path string, whitelist []string) error {
	// Thread safe Map of target:linkname
	var hardlinks sync.Map

	originalPerms := make([]OriginalPerm, 0)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return errors.Wrap(err, "Error getting next tar header")
		}
		target := filepath.Clean(filepath.Join(path, header.Name))
		// Make sure the target isn't part of the whitelist
		if checkWhitelist(target, whitelist) {
			continue
		}
		mode := header.FileInfo().Mode()
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); os.IsNotExist(err) {
				if mode.Perm()&(1<<(uint(7))) == 0 {
					logrus.Debugf("Write permission bit not set on %s by default; setting manually", target)
					originalMode := mode
					mode = mode | (1 << uint(7))
					// keep track of original file permission to reset later
					originalPerms = append(originalPerms, OriginalPerm{
						path: target,
						perm: originalMode,
					})
				}
				logrus.Debugf("Creating directory %s with permissions %v", target, mode)
				if err := os.MkdirAll(target, mode); err != nil {
					return err
				}
				// In some cases, MkdirAll doesn't change the permissions, so run Chmod
				if err := os.Chmod(target, mode); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			// It's possible for a file to be included before the directory it's in is created.
			baseDir := filepath.Dir(target)
			if _, err := os.Stat(baseDir); os.IsNotExist(err) {
				logrus.Debugf("baseDir %s for file %s does not exist. Creating", baseDir, target)
				if err := os.MkdirAll(baseDir, 0755); err != nil {
					return err
				}
			}
			// It's possible we end up creating files that can't be overwritten based on their permissions.
			// Explicitly delete an existing file before continuing.
			if _, err := os.Stat(target); !os.IsNotExist(err) {
				logrus.Debugf("Removing %s for overwrite", target)
				if err := os.Remove(target); err != nil {
					logrus.Errorf("error removing file %s", target)
					return err
				}
			}

			logrus.Debugf("Creating file %s with permissions %v", target, mode)
			currFile, err := os.Create(target)
			if err != nil {
				logrus.Errorf("Error creating file %s %s", target, err)
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
		case tar.TypeSymlink:
			// It's possible we end up creating files that can't be overwritten based on their permissions.
			// Explicitly delete an existing file before continuing.
			if _, err := os.Stat(target); !os.IsNotExist(err) {
				logrus.Debugf("Removing %s to create symlink", target)
				if err := os.RemoveAll(target); err != nil {
					logrus.Debugf("Unable to remove %s: %s", target, err)
				}
			}

			if err = os.Symlink(header.Linkname, target); err != nil {
				logrus.Errorf("Failed to create symlink between %s and %s: %s", header.Linkname, target, err)
			}
		case tar.TypeLink:
			linkname := filepath.Clean(filepath.Join(path, header.Linkname))
			// Check if the linkname already exists
			if _, err := os.Stat(linkname); !os.IsNotExist(err) {
				// If it exists, create the hard link
				resolveHardlink(linkname, target)
			} else {
				hardlinks.Store(target, linkname)
			}
		}
	}
	var resolveError atomic.Value
	hardlinks.Range(func(key, value interface{}) bool {
		target := key.(string)
		linkname := value.(string)
		logrus.Info("Resolving hard links")
		if _, err := os.Stat(linkname); !os.IsNotExist(err) {
			// If it exists, create the hard link
			if err := resolveHardlink(linkname, target); err != nil {
				resolveError.Store(errors.Wrap(err, fmt.Sprintf("Unable to create hard link from %s to %s", linkname, target)))
				return false
			}
		}
		return true
	})
	if resolveError.Load() != nil {
		return resolveError.Load().(error)
	}

	// reset all original file
	for _, perm := range originalPerms {
		if err := os.Chmod(perm.path, perm.perm); err != nil {
			return err
		}
	}
	return nil
}

func resolveHardlink(linkname, target string) error {
	if err := os.Link(linkname, target); err != nil {
		return err
	}
	logrus.Debugf("Created hard link from %s to %s", linkname, target)
	return nil
}

func checkWhitelist(target string, whitelist []string) bool {
	for _, w := range whitelist {
		if HasFilepathPrefix(target, w) {
			logrus.Debugf("Not extracting %s, as it has prefix %s which is whitelisted", target, w)
			return true
		}
	}
	return false
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
