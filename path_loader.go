package mjingo

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// PathLoader is a helper to load templates from a given directory.
//
// This creates a dynamic loader which looks up templates in the
// given directory.  Templates that start with a dot (`.`) or are contained in
// a folder starting with a dot cannot be loaded.
func PathLoader(dir string) LoadFunc {
	return func(name string) (string, error) {
		if !isSafeRelPath(name) {
			return "", NewErrorNotFound(name)
		}
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return "", NewErrorNotFound(name)
			}
			return "", err
		}
		return string(data), nil
	}
}

func isSafeRelPath(relPath string) bool {
	for _, segment := range strings.Split(relPath, "/") {
		if strings.HasPrefix(segment, ".") || strings.Contains(segment, "\\") {
			return false
		}
	}
	return true
}
