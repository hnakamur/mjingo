package mjingo

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplate(t *testing.T) {
	refFilenames := mustGlob(t, []string{"tests", "inputs", "refs"}, []string{"*.txt", "*.html"})
	refContents := make(map[string]string)
	for _, refFilename := range refFilenames {
		refFileBasename := filepath.Base(refFilename)
		refContent := mustReadFile(t, refFilename)
		refContents[refFileBasename] = refContent
	}

	inputFilenames := mustGlob(t, []string{"tests", "inputs"}, []string{"*.txt", "*.html"})
	for _, inputFilename := range inputFilenames {
		inputFileBasename := filepath.Base(inputFilename)
		t.Run(inputFileBasename, func(t *testing.T) {
			inputContent := mustReadFile(t, inputFilename)
			jsonContent, templateContent, found := strings.Cut(inputContent, "\n---\n")
			if !found {
				t.Fatalf(`input file does not contain "\n---\n" separator, file=%s`, inputFilename)
			}
			var ctx any
			decoder := json.NewDecoder(strings.NewReader(jsonContent))
			decoder.UseNumber()
			if err := decoder.Decode(&ctx); err != nil {
				t.Fatalf("cannot decode JSON in file=%s, err=%s", inputFilename, err)
			}

			env := NewEnvironment()
			env.SetDebug(true)
			for refBaseFilename, refContent := range refContents {
				if err := env.AddTemplate(refBaseFilename, refContent); err != nil {
					t.Fatal(err)
				}
			}
			var got string
			err := env.AddTemplate(inputFileBasename, templateContent)
			if err != nil {
				got = err.Error()
			} else {
				tpl, err := env.GetTemplate(inputFileBasename)
				if err != nil {
					t.Fatal(err)
				}
				got, err = tpl.Render(ValueFromGoValue(ctx))
				if err != nil {
					var b strings.Builder
					fmt.Fprintf(&b, "!!!ERROR!!!\n\n%#q\n\n", err)
					fmt.Fprintf(&b, "%#s", err)
					for merr := (*Error)(nil); errors.As(err, &merr) && merr.source != nil; err = merr.source {
						fmt.Fprintf(&b, "\ncaused by: %#s", merr.source)
					}
					got = b.String()
				}
			}
			checkResultWithSnapshotFile(t, got, inputFilename)
		})
	}
}

func TestVMBlockFragments(t *testing.T) {
	refFilenames := mustGlob(t, []string{"tests", "fragment-inputs", "refs"}, []string{"*.txt"})
	refContents := make(map[string]string)
	for _, refFilename := range refFilenames {
		refFileBasename := filepath.Base(refFilename)
		refContent := mustReadFile(t, refFilename)
		refContents[refFileBasename] = refContent
	}

	inputFilenames := mustGlob(t, []string{"tests", "fragment-inputs"}, []string{"*.txt"})
	for _, inputFilename := range inputFilenames {
		inputFileBasename := filepath.Base(inputFilename)
		t.Run(inputFileBasename, func(t *testing.T) {
			inputContent := mustReadFile(t, inputFilename)
			jsonContent, templateContent, found := strings.Cut(inputContent, "\n---\n")
			if !found {
				t.Fatalf(`input file does not contain "\n---\n" separator, file=%s`, inputFilename)
			}
			var ctx any
			decoder := json.NewDecoder(strings.NewReader(jsonContent))
			decoder.UseNumber()
			if err := decoder.Decode(&ctx); err != nil {
				t.Fatalf("cannot decode JSON in file=%s, err=%s", inputFilename, err)
			}

			env := NewEnvironment()
			for refBaseFilename, refContent := range refContents {
				if err := env.AddTemplate(refBaseFilename, refContent); err != nil {
					t.Fatal(err)
				}
			}
			var got string
			err := env.AddTemplate(inputFilename, templateContent)
			if err != nil {
				got = err.Error()
			} else {
				tpl, err := env.GetTemplate(inputFilename)
				if err != nil {
					t.Fatal(err)
				}
				state, err := tpl.EvalToState(ValueFromGoValue(ctx))
				if err != nil {
					got = err.Error()
				}
				rendered, err := state.RenderBlock("fragment")
				if err != nil {
					got = err.Error()
				} else {
					got = rendered + "\n"
				}
			}
			checkResultWithSnapshotFile(t, got, inputFilename)
		})
	}
}
