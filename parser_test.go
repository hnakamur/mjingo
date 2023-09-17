package mjingo

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"
)

func TestParser(t *testing.T) {
	debugStringStmt := func(stmt statement) string {
		return fmt.Sprintf("Ok(\n" +
			")")
	}

	debugStringErr := func(err error) string {
		var merr *Error
		if errors.As(err, &merr) {
			return fmt.Sprintf("Err(\n"+
				"    Error {\n"+
				"        kind: %s,\n"+
				"        detail: %q,\n"+
				"        name: %q,\n"+
				"        line: %d,\n"+
				"    },\n"+
				")\n", merr.Kind().debugString(), merr.detail, merr.name.Unwrap(), merr.lineno)
		}
		return fmt.Sprintf("Err(\n"+
			"    %s\n"+
			")\n", err)
	}

	inputFilenames := mustGlob(t, []string{"tests", "parser-inputs"}, []string{"*.txt"})
	for _, inputFilename := range inputFilenames {
		t.Run(inputFilename, func(t *testing.T) {
			inputContent := mustReadFile(t, inputFilename)
			ast, err := parse(inputContent, filepath.Base(inputFilename))
			var got string
			if err != nil {
				got = debugStringErr(err)
			} else {
				got = debugStringStmt(ast)
			}
			checkResultWithSnapshotFile(t, got, inputFilename)
		})
	}
}
