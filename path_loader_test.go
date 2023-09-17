package mjingo_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/hnakamur/mjingo"
)

func TestPathLoader(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "hello.j2"), []byte("Hello {{ name }}"), 0o600); err != nil {
			t.Fatal(err)
		}
		env := mjingo.NewEnvironment()
		env.SetLoader(mjingo.PathLoader(dir))
		tpl, err := env.GetTemplate("hello.j2")
		if err != nil {
			t.Fatal(err)
		}
		ctx := mjingo.ValueFromGoValue(map[string]string{"name": "John"})
		got, err := tpl.Render(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if want := "Hello John"; got != want {
			t.Errorf("result mismatch, got=%q, want=%q", got, want)
		}
	})
	t.Run("notFound", func(t *testing.T) {
		dir := t.TempDir()
		env := mjingo.NewEnvironment()
		env.SetLoader(mjingo.PathLoader(dir))
		_, err := env.GetTemplate("no_such_template.j2")
		if err == nil {
			t.Errorf("should get error but not")
		}
		if merr := (*mjingo.Error)(nil); errors.As(err, &merr) {
			if got, want := merr.Error(), "template not found: template no_such_template.j2 does not exist"; got != want {
				t.Errorf("error message mismatch, got=%q, want=%q", got, want)
			}
		} else {
			t.Errorf("error type mismatch, got=%T, want=%T", err, mjingo.Error{})
		}
	})
	t.Run("unsafePath", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".hello.j2"), []byte("Hello {{ name }}"), 0o600); err != nil {
			t.Fatal(err)
		}
		env := mjingo.NewEnvironment()
		env.SetLoader(mjingo.PathLoader(dir))
		_, err := env.GetTemplate(".hello.j2")
		if merr := (*mjingo.Error)(nil); errors.As(err, &merr) {
			if got, want := merr.Error(), "template not found: template .hello.j2 does not exist"; got != want {
				t.Errorf("error message mismatch, got=%q, want=%q", got, want)
			}
		} else {
			t.Errorf("error type mismatch, got=%T, want=%T", err, mjingo.Error{})
		}
	})
}
