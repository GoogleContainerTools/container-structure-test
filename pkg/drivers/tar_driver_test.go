package drivers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleContainerTools/container-structure-test/internal/pkgutil"
)

func TestTarDriverFileOperationsStayWithinImageRoot(t *testing.T) {
	parent := t.TempDir()
	imageRoot := filepath.Join(parent, "image-root")
	if err := os.MkdirAll(filepath.Join(imageRoot, "etc"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(imageRoot, "etc", "config.txt"), []byte("image-data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "outside.txt"), []byte("outside-data"), 0644); err != nil {
		t.Fatal(err)
	}

	driver := &TarDriver{Image: pkgutil.Image{FSPath: imageRoot}}

	if got, err := driver.ReadFile("/etc/config.txt"); err != nil || string(got) != "image-data" {
		t.Fatalf("ReadFile inside image root = %q, %v", got, err)
	}
	if _, err := driver.StatFile("/etc/config.txt"); err != nil {
		t.Fatalf("StatFile inside image root: %v", err)
	}
	if _, err := driver.ReadDir("/etc"); err != nil {
		t.Fatalf("ReadDir inside image root: %v", err)
	}

	for name, op := range map[string]func() error{
		"ReadFile": func() error {
			_, err := driver.ReadFile("../outside.txt")
			return err
		},
		"StatFile": func() error {
			_, err := driver.StatFile("../outside.txt")
			return err
		},
		"ReadDir": func() error {
			_, err := driver.ReadDir("..")
			return err
		},
	} {
		if err := op(); err == nil {
			t.Fatalf("%s allowed a path outside the image root", name)
		}
	}
}

func TestTarDriverResolvesSymlinksWithinImageRoot(t *testing.T) {
	imageRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(imageRoot, "etc"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(imageRoot, "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(imageRoot, "etc", "config.txt"), []byte("image-data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/etc/config.txt", filepath.Join(imageRoot, "absolute-link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../etc/config.txt", filepath.Join(imageRoot, "bin", "relative-link")); err != nil {
		t.Fatal(err)
	}

	driver := &TarDriver{Image: pkgutil.Image{FSPath: imageRoot}}

	for _, imagePath := range []string{"/absolute-link", "/bin/relative-link"} {
		got, err := driver.ReadFile(imagePath)
		if err != nil {
			t.Fatalf("ReadFile(%q): %v", imagePath, err)
		}
		if string(got) != "image-data" {
			t.Fatalf("ReadFile(%q) = %q", imagePath, got)
		}
	}

	info, err := driver.StatFile("/absolute-link")
	if err != nil {
		t.Fatalf("StatFile on symlink: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("StatFile should return the final symlink itself, got mode %s", info.Mode())
	}
}

func TestTarDriverRejectsSymlinksOutsideImageRoot(t *testing.T) {
	parent := t.TempDir()
	imageRoot := filepath.Join(parent, "image-root")
	if err := os.MkdirAll(imageRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "outside.txt"), []byte("outside-data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../outside.txt", filepath.Join(imageRoot, "file-link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("..", filepath.Join(imageRoot, "dir-link")); err != nil {
		t.Fatal(err)
	}

	driver := &TarDriver{Image: pkgutil.Image{FSPath: imageRoot}}

	if _, err := driver.ReadFile("/file-link"); err == nil {
		t.Fatal("ReadFile allowed a symlink outside the image root")
	}
	if _, err := driver.ReadDir("/dir-link"); err == nil {
		t.Fatal("ReadDir allowed a symlink outside the image root")
	}
	if _, err := driver.StatFile("/dir-link/outside.txt"); err == nil {
		t.Fatal("StatFile allowed an intermediate symlink outside the image root")
	}
}
