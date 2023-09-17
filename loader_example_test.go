package mjingo_test

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hnakamur/mjingo"
)

func MyPathLoader(dir string) mjingo.LoadFunc {
	return func(name string) (string, error) {
		if os.PathSeparator != '/' {
			name = strings.Join(strings.Split(name, "/"), string(os.PathSeparator))
		}
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return "", mjingo.NewErrorNotFound(name)
			}
			return "", err
		}
		return string(data), nil
	}
}

func ExampleLoadFunc() {
	dir, err := os.MkdirTemp("", "mjingo-test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	const templateSubdir = "subdir"
	const templateName = templateSubdir + "/hello.j2"
	if err := os.MkdirAll(filepath.Join(dir, templateSubdir), 0o700); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, templateName), []byte("Hello {{ name }}"), 0o600); err != nil {
		log.Fatal(err)
	}

	env := mjingo.NewEnvironment()
	env.SetLoader(MyPathLoader(dir))
	tpl, err := env.GetTemplate(templateName)
	if err != nil {
		log.Fatal(err)
	}
	ctx := mjingo.ValueFromGoValue(map[string]string{"name": "John"})
	got, err := tpl.Render(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(got)
	// Output: Hello John
}
