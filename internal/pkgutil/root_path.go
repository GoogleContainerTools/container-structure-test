package pkgutil

import (
	"fmt"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"
)

const maxRootPathSymlinks = 255

// ResolvePathInRoot resolves imagePath below root. Absolute paths are treated as
// absolute within root, and symlink targets are resolved using the same image
// root instead of the host root.
func ResolvePathInRoot(root, imagePath string, followFinalSymlink bool) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absRoot = filepath.Clean(absRoot)

	parts, err := resolveRootPathParts(nil, imagePath)
	if err != nil {
		return "", err
	}

	resolved := []string{}
	symlinkCount := 0
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		candidate := joinRootPath(absRoot, append(resolved, part))
		isFinal := i == len(parts)-1
		if !isFinal || followFinalSymlink {
			info, err := os.Lstat(candidate)
			if err != nil && !os.IsNotExist(err) {
				return "", err
			}
			if err == nil && info.Mode()&os.ModeSymlink != 0 {
				symlinkCount++
				if symlinkCount > maxRootPathSymlinks {
					return "", fmt.Errorf("too many symlinks resolving path %q", imagePath)
				}
				linkTarget, err := os.Readlink(candidate)
				if err != nil {
					return "", err
				}
				base := resolved
				if pathpkg.IsAbs(filepath.ToSlash(linkTarget)) {
					base = nil
				}
				linkParts, err := resolveRootPathParts(base, linkTarget)
				if err != nil {
					return "", err
				}
				parts = append(linkParts, parts[i+1:]...)
				resolved = nil
				i = -1
				continue
			}
		}
		resolved = append(resolved, part)
	}
	return joinRootPath(absRoot, resolved), nil
}

func resolveRootPathParts(base []string, imagePath string) ([]string, error) {
	cleaned := filepath.ToSlash(imagePath)
	if pathpkg.IsAbs(cleaned) {
		base = nil
	}

	parts := append([]string{}, base...)
	for _, part := range strings.Split(pathpkg.Clean(cleaned), "/") {
		switch part {
		case "", ".":
			continue
		case "..":
			if len(parts) == 0 {
				return nil, fmt.Errorf("path %q escapes root", imagePath)
			}
			parts = parts[:len(parts)-1]
		default:
			parts = append(parts, part)
		}
	}
	return parts, nil
}

func joinRootPath(root string, parts []string) string {
	return filepath.Join(append([]string{root}, parts...)...)
}
