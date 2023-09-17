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
//
// The name argument of the returned LoadFunc can contain `/` as a path separator
// (even on Windows).
// If name contains `\`, an [Error] with [TemplateNotFound] kind will be returned
// from the returned LoadFunc.
func PathLoader(dir string) LoadFunc {
	return func(name string) (string, error) {
		segments := strings.Split(name, "/")
		for _, segment := range segments {
			if strings.HasPrefix(segment, ".") || strings.Contains(segment, `\`) {
				return "", NewErrorNotFound(name)
			}
		}
		if os.PathSeparator != '/' {
			name = strings.Join(segments, string(os.PathSeparator))
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
