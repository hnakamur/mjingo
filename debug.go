package mjingo

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/hnakamur/mjingo/option"
)

type debugInfo struct {
	templateSource   string
	referencedLocals map[string]Value
}

type varPrinter map[string]Value

func (p varPrinter) Format(f fmt.State, verb rune) {
	switch verb {
	case 's', 'q':
		if len(p) == 0 {
			io.WriteString(f, "No referenced variables")
			return
		}
		s := newDebugStruct("Referenced variables:")
		for _, key := range mapSortedKeys(p) {
			s.field(key, p[key])
		}
		s.Format(f, verb)
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods varPrinter
		type varPrinter hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), varPrinter(p))
	}
}

func (d debugInfo) render(w io.Writer, name option.Option[string], kind ErrorKind,
	line option.Option[uint], spn option.Option[span]) error {
	if len(d.templateSource) > 0 {
		title := name.UnwrapOr("")
		if len(title) > 0 {
			if pos := strings.LastIndexAny(title, `/\`); pos != -1 {
				title = title[pos+1:]
			}
		}
		if len(title) == 0 {
			title = "Template Source"
		}
		title = fmt.Sprintf(" %s ", title)
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if err := writeCenterAligned(w, title, '-', 79); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		lines, err := splitSourceLines(d.templateSource)
		if err != nil {
			return err
		}
		idx := uintSaturatingSub(line.UnwrapOr(1), 1)
		skip := uintSaturatingSub(idx, 3)
		for i := skip; i < min(idx, 3); i++ {
			if _, err := fmt.Fprintf(w, "%4d | %s\n", i+1, lines[i]); err != nil {
				return err
			}
		}
		if idx < uint(len(lines)) {
			if _, err := fmt.Fprintf(w, "%4d > %s\n", idx+1, lines[idx]); err != nil {
				return err
			}
		}
		if sp := (span{}); spn.UnwrapTo(&sp) && sp.StartLine == sp.EndLine {
			if _, err := fmt.Fprintf(w,
				"     i %s%s %s\n",
				strings.Repeat(" ", int(sp.StartCol)),
				strings.Repeat("^", int(sp.EndCol-sp.StartCol)),
				kind); err != nil {
				return err
			}

		}
		for i := idx + 1; i < min(idx+3, uint(len(lines))); i++ {
			if _, err := fmt.Fprintf(w, "%4d | %s\n", i+1, lines[i]); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, strings.Repeat("~", 79)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%q\n", varPrinter(d.referencedLocals)); err != nil {
		return err
	}
	if _, err := io.WriteString(w, strings.Repeat("-", 79)); err != nil {
		return err
	}
	return nil
}

type debugStruct struct {
	name        string
	indentLevel uint
	fields      []debugStructField
}

type debugStructField struct {
	name  string
	value any
}

func newDebugStruct(name string) *debugStruct {
	return &debugStruct{name: name}
}

func (s *debugStruct) field(name string, value any) *debugStruct {
	s.fields = append(s.fields, debugStructField{name: name, value: value})
	return s
}

func (s debugStruct) Format(f fmt.State, verb rune) {
	switch verb {
	case 's':
		io.WriteString(f, s.name)
		io.WriteString(f, " { ")
		for i, field := range s.fields {
			if i > 0 {
				io.WriteString(f, ", ")
			}
			io.WriteString(f, field.name)
			io.WriteString(f, ": ")
			fmt.Fprintf(f, "%v", field.value)
		}
		io.WriteString(f, " }")
	case 'q':
		io.WriteString(f, s.name)
		io.WriteString(f, " {\n")
		s.indentLevel++
		for _, field := range s.fields {
			writeIndent(f, s.indentLevel)
			io.WriteString(f, field.name)
			io.WriteString(f, ": ")
			fmt.Fprintf(f, "%v", field.value)
			io.WriteString(f, ",\n")
		}
		s.indentLevel--
		writeIndent(f, s.indentLevel)
		io.WriteString(f, "}")
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods debugStruct
		type debugStruct hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), debugStruct(s))
	}
}

func writeIndent(w io.Writer, level uint) error {
	for i := uint(0); i < level; i++ {
		if _, err := io.WriteString(w, "    "); err != nil {
			return err
		}
	}
	return nil
}

func writeCenterAligned(w io.Writer, s string, filler rune, width uint) error {
	nLeftRight := uintSaturatingSub(width, uint(len(s)))
	nLeft := nLeftRight / 2
	nRight := uintSaturatingSub(nLeftRight, nLeft)
	if nLeft > 0 {
		if _, err := io.WriteString(w, strings.Repeat(string(filler), int(nLeft))); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, s); err != nil {
		return err
	}
	if nRight > 0 {
		if _, err := io.WriteString(w, strings.Repeat(string(filler), int(nRight))); err != nil {
			return err
		}
	}
	return nil
}

func formatCenterAligned(s string, filler rune, width uint) string {
	var b strings.Builder
	_ = writeCenterAligned(&b, s, filler, width)
	return b.String()
}

func splitSourceLines(source string) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(source))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
