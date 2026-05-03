package pkgutil

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func testTarReader(t *testing.T, entries ...func(*tar.Writer) error) *tar.Reader {
	t.Helper()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, entry := range entries {
		if err := entry(tw); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	return tar.NewReader(bytes.NewReader(buf.Bytes()))
}

func testTarFile(name, body string) func(*tar.Writer) error {
	return func(tw *tar.Writer) error {
		if err := tw.WriteHeader(&tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(body)),
		}); err != nil {
			return err
		}
		_, err := tw.Write([]byte(body))
		return err
	}
}

func testTarDir(name string) func(*tar.Writer) error {
	return func(tw *tar.Writer) error {
		return tw.WriteHeader(&tar.Header{
			Name:     name,
			Mode:     0755,
			Typeflag: tar.TypeDir,
		})
	}
}

func testTarSymlink(name, linkname string) func(*tar.Writer) error {
	return func(tw *tar.Writer) error {
		return tw.WriteHeader(&tar.Header{
			Name:     name,
			Linkname: linkname,
			Typeflag: tar.TypeSymlink,
		})
	}
}

func testTarHardlink(name, linkname string) func(*tar.Writer) error {
	return func(tw *tar.Writer) error {
		return tw.WriteHeader(&tar.Header{
			Name:     name,
			Linkname: linkname,
			Typeflag: tar.TypeLink,
		})
	}
}

func TestUnpackTarRejectsEntriesOutsideRoot(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}

	err := unpackTar(testTarReader(t, testTarFile("../outside.txt", "outside-data")), root, nil)
	if err == nil {
		t.Fatal("unpackTar allowed an entry outside the root")
	}
	if _, err := os.Stat(filepath.Join(parent, "outside.txt")); !os.IsNotExist(err) {
		t.Fatalf("outside path exists after rejected unpack: %v", err)
	}
}

func TestUnpackTarRejectsSymlinksOutsideRoot(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}

	err := unpackTar(testTarReader(t, testTarSymlink("link", "..")), root, nil)
	if err == nil {
		t.Fatal("unpackTar allowed a symlink outside the root")
	}
}

func TestUnpackTarRejectsHardlinksOutsideRoot(t *testing.T) {
	root := t.TempDir()

	err := unpackTar(testTarReader(t, testTarHardlink("link", "../outside.txt")), root, nil)
	if err == nil {
		t.Fatal("unpackTar allowed a hard link outside the root")
	}
}

func TestUnpackTarHardlinkToSymlinkDoesNotFollowTarget(t *testing.T) {
	root := t.TempDir()

	err := unpackTar(testTarReader(t,
		testTarFile("file.txt", "image-data"),
		testTarSymlink("sym", "file.txt"),
		testTarHardlink("hard", "sym"),
	), root, nil)
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Lstat(filepath.Join(root, "hard"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("hard link to symlink should remain a symlink, got mode %s", info.Mode())
	}
	target, err := os.Readlink(filepath.Join(root, "hard"))
	if err != nil {
		t.Fatal(err)
	}
	if target != "file.txt" {
		t.Fatalf("hard link to symlink target = %q", target)
	}
}

func TestUnpackTarRejectsHardlinkSourceSymlinkOutsideRoot(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "outside.txt"), []byte("outside-data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../outside.txt", filepath.Join(root, "sym")); err != nil {
		t.Fatal(err)
	}

	err := unpackTar(testTarReader(t, testTarHardlink("hard", "sym")), root, nil)
	if err == nil {
		t.Fatal("unpackTar allowed a hard link through a symlink outside the root")
	}
	if _, err := os.Lstat(filepath.Join(root, "hard")); !os.IsNotExist(err) {
		t.Fatalf("hard link path exists after rejected unpack: %v", err)
	}
}

func TestUnpackTarResolvesSymlinksWithinRoot(t *testing.T) {
	root := t.TempDir()

	err := unpackTar(testTarReader(t,
		testTarDir("etc"),
		testTarSymlink("link", "/etc"),
		testTarFile("link/generated.txt", "image-data"),
	), root, nil)
	if err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(root, "etc", "generated.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "image-data" {
		t.Fatalf("unpacked file through image symlink = %q", got)
	}
}
