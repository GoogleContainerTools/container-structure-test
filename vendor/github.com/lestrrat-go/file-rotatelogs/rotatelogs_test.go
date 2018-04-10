package rotatelogs_test

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/stretchr/testify/assert"
)

func TestSatisfiesIOWriter(t *testing.T) {
	var w io.Writer
	w, _ = rotatelogs.New("/foo/bar")
	_ = w
}

func TestSatisfiesIOCloser(t *testing.T) {
	var c io.Closer
	c, _ = rotatelogs.New("/foo/bar")
	_ = c
}

func TestLogRotate(t *testing.T) {
	dir, err := ioutil.TempDir("", "file-rotatelogs-test")
	if !assert.NoError(t, err, "creating temporary directory should succeed") {
		return
	}
	defer os.RemoveAll(dir)

	// Change current time, so we can safely purge old logs
	dummyTime := time.Now().Add(-7 * 24 * time.Hour)
	dummyTime = dummyTime.Add(time.Duration(-1 * dummyTime.Nanosecond()))
	clock := clockwork.NewFakeClockAt(dummyTime)
	linkName := filepath.Join(dir, "log")
	rl, err := rotatelogs.New(
		filepath.Join(dir, "log%Y%m%d%H%M%S"),
		rotatelogs.WithClock(clock),
		rotatelogs.WithMaxAge(24*time.Hour),
		rotatelogs.WithLinkName(linkName),
	)
	if !assert.NoError(t, err, `rotatelogs.New should succeed`) {
		return
	}
	defer rl.Close()

	str := "Hello, World"
	n, err := rl.Write([]byte(str))
	if !assert.NoError(t, err, "rl.Write should succeed") {
		return
	}

	if !assert.Len(t, str, n, "rl.Write should succeed") {
		return
	}

	fn := rl.CurrentFileName()
	if fn == "" {
		t.Errorf("Could not get filename %s", fn)
	}

	content, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Errorf("Failed to read file %s: %s", fn, err)
	}

	if string(content) != str {
		t.Errorf(`File content does not match (was "%s")`, content)
	}

	err = os.Chtimes(fn, dummyTime, dummyTime)
	if err != nil {
		t.Errorf("Failed to change access/modification times for %s: %s", fn, err)
	}

	fi, err := os.Stat(fn)
	if err != nil {
		t.Errorf("Failed to stat %s: %s", fn, err)
	}

	if !fi.ModTime().Equal(dummyTime) {
		t.Errorf("Failed to chtime for %s (expected %s, got %s)", fn, fi.ModTime(), dummyTime)
	}

	clock.Advance(time.Duration(7 * 24 * time.Hour))

	// This next Write() should trigger Rotate()
	rl.Write([]byte(str))
	newfn := rl.CurrentFileName()
	if newfn == fn {
		t.Errorf(`New file name and old file name should not match ("%s" != "%s")`, fn, newfn)
	}

	content, err = ioutil.ReadFile(newfn)
	if err != nil {
		t.Errorf("Failed to read file %s: %s", newfn, err)
	}

	if string(content) != str {
		t.Errorf(`File content does not match (was "%s")`, content)
	}

	time.Sleep(time.Second)

	// fn was declared above, before mocking CurrentTime
	// Old files should have been unlinked
	_, err = os.Stat(fn)
	if !assert.Error(t, err, "os.Stat should have failed") {
		return
	}

	linkDest, err := os.Readlink(linkName)
	if err != nil {
		t.Errorf("Failed to readlink %s: %s", linkName, err)
	}

	if linkDest != newfn {
		t.Errorf(`Symlink destination does not match expected filename ("%s" != "%s")`, newfn, linkDest)
	}
}

func TestLogSetOutput(t *testing.T) {
	dir, err := ioutil.TempDir("", "file-rotatelogs-test")
	if err != nil {
		t.Errorf("Failed to create temporary directory: %s", err)
	}
	defer os.RemoveAll(dir)

	rl, err := rotatelogs.New(filepath.Join(dir, "log%Y%m%d%H%M%S"))
	if !assert.NoError(t, err, `rotatelogs.New should succeed`) {
		return
	}
	defer rl.Close()

	log.SetOutput(rl)
	defer log.SetOutput(os.Stderr)

	str := "Hello, World"
	log.Print(str)

	fn := rl.CurrentFileName()
	if fn == "" {
		t.Errorf("Could not get filename %s", fn)
	}

	content, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Errorf("Failed to read file %s: %s", fn, err)
	}

	if !strings.Contains(string(content), str) {
		t.Errorf(`File content does not contain "%s" (was "%s")`, str, content)
	}
}
