package mjingo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCompiler(t *testing.T) {
	t.Run("for_loop", func(t *testing.T) {
		c := newCodeGenerator("<unknown>", "")
		c.add(lookupInstruction{Name: "items"})
		c.startForLoop(true, false)
		c.add(emitInstruction{})
		c.endForLoop(false)
		c.add(emitRawInstruction{Val: "!"})
		insts, blocks := c.finish()
		testVerifyInstsAndBlocksWithSnapshot(t, insts, blocks,
			filepath.Join("tests", "compiler", "for_loop.snap"))
	})
	t.Run("if_branches", func(t *testing.T) {
		c := newCodeGenerator("<unknown>", "")
		c.add(lookupInstruction{Name: "false"})
		c.startIf()
		c.add(emitRawInstruction{Val: "nope1"})
		c.startElse()
		c.add(lookupInstruction{Name: "nil"})
		c.startIf()
		c.add(emitRawInstruction{Val: "nope1"})
		c.startElse()
		c.add(emitRawInstruction{Val: "yes"})
		c.endIf()
		c.endIf()
		insts, blocks := c.finish()
		testVerifyInstsAndBlocksWithSnapshot(t, insts, blocks,
			filepath.Join("tests", "compiler", "if_branches.snap"))
	})
	t.Run("bool_ops", func(t *testing.T) {
		c := newCodeGenerator("<unknown>", "")
		c.startScBool()
		c.add(lookupInstruction{Name: "first"})
		c.scBool(true)
		c.add(lookupInstruction{Name: "second"})
		c.scBool(false)
		c.add(lookupInstruction{Name: "third"})
		c.endScBool()
		insts, blocks := c.finish()
		testVerifyInstsAndBlocksWithSnapshot(t, insts, blocks,
			filepath.Join("tests", "compiler", "bool_ops.snap"))
	})
	t.Run("const", func(t *testing.T) {
		c := newCodeGenerator("<unknown>", "")
		c.add(loadConstInstruction{Val: ValueFromGoValue("a")})
		c.add(loadConstInstruction{Val: ValueFromGoValue(42)})
		c.add(stringConcatInstruction{})
		insts, blocks := c.finish()
		testVerifyInstsAndBlocksWithSnapshot(t, insts, blocks,
			filepath.Join("tests", "compiler", "const.snap"))
	})
	t.Run("referencedNamesEmptyBug", func(t *testing.T) {
		c := newCodeGenerator("<unknown>", "")
		insts, _ := c.finish()
		names := insts.getReferencedNames(0)
		if got, want := len(names), 0; got != want {
			t.Errorf("name count mismatch, got=%d, want=%d", got, want)
		}
	})
}

func testVerifyInstsAndBlocksWithSnapshot(t *testing.T, insts instructions,
	blocks map[string]instructions, snapFilename string) {
	t.Helper()
	got := debugStringInstsAndBlocks(insts, blocks)
	want := mustReadFile(t, snapFilename)
	if got != want {
		t.Errorf("result mismatch\n-- got -- \n%s\n-- want --\n%s\n-- diff --\n%s",
			got, want, cmp.Diff(got, want))
		if overwriteSnapshot {
			if err := os.WriteFile(snapFilename, []byte(got), 0o644); err != nil {
				t.Fatal(err)
			}
			t.Logf("overwritten test snapshot file: %s", snapFilename)
		} else {
			t.Logf("If `got` result is correct, rerun tests with -overwrite-snapshot flag to overwrite snapshot file")
		}
	}
}

func debugStringInstsAndBlocks(insts instructions,
	blocks map[string]instructions) string {

	const indent1 = "    "

	debugPrintInsts := func(w io.Writer, insts instructions, prefix, indent string) {
		if len(insts.instructions) == 0 {
			fmt.Fprintf(w, "%s[],\n", prefix)
			return
		}
		fmt.Fprintf(w, "%s[\n", prefix)
		j := 0
		for i, inst := range insts.instructions {
			fmt.Fprintf(w, "%s%s%05d | %s", indent, indent1, i, inst)
			if i == 0 || (j < len(insts.lineInfos) && insts.lineInfos[j].firstInstruction == uint32(i)) {
				fmt.Fprintf(w, "  [line %d],\n", insts.lineInfos[j].line)
				j++
			} else {
				fmt.Fprint(w, ",\n")
			}
		}
		fmt.Fprintf(w, "%s],\n", indent)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "(\n")
	debugPrintInsts(&b, insts, indent1, indent1)
	if len(blocks) == 0 {
		fmt.Fprintf(&b, "%s{},\n", indent1)
	} else {
		keys := make([]string, 0, len(blocks))
		for key := range blocks {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		fmt.Fprintf(&b, "%s{\n", indent1)
		nextIndent := indent1 + indent1
		for _, key := range keys {
			fmt.Fprintf(&b, "%s%s%q: ", indent1, indent1, key)
			debugPrintInsts(&b, blocks[key], "", nextIndent)
		}
		fmt.Fprintf(&b, "%s},\n", indent1)
	}
	fmt.Fprintf(&b, ")\n")
	return b.String()
}
