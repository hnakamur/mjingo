package mjingo

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplate(t *testing.T) {
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
			env.SetKeepTrailingNewline(true)
			var got string
			err := env.AddTemplate(inputFilename, templateContent)
			if err != nil {
				got = err.Error()
			} else {
				tpl, err := env.GetTemplate(inputFilename)
				if err != nil {
					t.Fatal(err)
				}
				got, err = tpl.Render(ValueFromGoValue(ctx))
				if err != nil {
					got = err.Error()
				}
			}
			checkResultWithSnapshotFile(t, got, inputFilename)
		})
	}
}
