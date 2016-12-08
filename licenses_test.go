package main

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

// Not currently used, but leaving the possibility open
type LicenseTestv1 struct {
	Debian bool
	Files  []string
}

var (
	// Whitelist is the list of packages that we want to automatically pass this
	// check even if it would normally fail for one reason or another.
	whitelist = []string{}

	// Blacklist is the set of words that, if contained in a license file, should cause a failure.
	// This will most likely just be names of unsupported licenses.
	blacklist = []string{"AGPL", "WTFPL"}
)

func checkFile(t *testing.T, licenseFile string) {
	// Read through the copyright file and make sure don't have an unauthorized license
	license, err := ioutil.ReadFile(licenseFile)
	if err != nil {
		t.Errorf("Error reading license file for %s: %s", licenseFile, err.Error())
		return
	}
	contents := strings.ToUpper(string(license))
	for _, b := range blacklist {
		if strings.Contains(contents, b) {
			t.Errorf("Invalid license for %s, license contains %s", licenseFile, b)
			return
		}
	}
}

func checkLicenses(t *testing.T, tt LicenseTestv1) {
	if tt.Debian {
		root := "/usr/share/doc"
		packages, err := ioutil.ReadDir(root)
		if err != nil {
			t.Fatalf("%s", err)
		}
		for _, p := range packages {
			if !p.IsDir() {
				continue
			}

			t.Logf(p.Name())

			// Skip over packages in the whitelist
			whitelisted := false
			for _, w := range whitelist {
				if w == p.Name() {
					whitelisted = true
					break
				}
			}
			if whitelisted {
				continue
			}

			// If package doesn't have copyright file, log an error.
			licenseFile := path.Join(root, p.Name(), "copyright")
			_, err := os.Stat(licenseFile)
			if err != nil {
				t.Errorf("Error reading license file for %s: %s", p.Name(), err.Error())
				continue
			}

			checkFile(t, licenseFile)
		}
	}

	for _, file := range tt.Files {
		checkFile(t, file)
	}
}
