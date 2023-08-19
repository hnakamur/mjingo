package vm

import (
	"testing"

	"github.com/hnakamur/mjingo/value"
)

func TestEnvironment(t *testing.T) {
	testCases := []struct {
		name    string
		source  string
		context any
		want    string
	}{
		{
			name:   "var",
			source: "Hello {{ name }}",
			context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
				Key: value.KeyRefFromString("name"), Value: value.FromString("World"),
			}})),
			want: "Hello World",
		},
		{
			name:    "stringExpr",
			source:  `Hello {{ "world" }}`,
			context: value.None,
			want:    "Hello world",
		},
		{
			name:    "i64Expr",
			source:  `Hello {{ 3 }}`,
			context: value.None,
			want:    "Hello 3",
		},
		{
			name:    "f64Expr",
			source:  `Hello {{ 3.14 }}`,
			context: value.None,
			want:    "Hello 3.14",
		},
		{
			name:    "boolExprTrue",
			source:  `Hello {{ true }}`,
			context: value.None,
			want:    "Hello true",
		},
		{
			name:    "boolExprFalse",
			source:  `Hello {{ False }}`,
			context: value.None,
			want:    "Hello false",
		},
		{
			name:    "noneExpr",
			source:  `Hello {{ none }}`,
			context: value.None,
			want:    "Hello none",
		},
		{
			name:   "getFastAttr",
			source: `Hello {{ user.name }}`,
			context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
				Key: value.KeyRefFromString("user"),
				Value: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
					Key:   value.KeyRefFromString("name"),
					Value: value.FromString("John"),
				}})),
			}})),
			want: "Hello John",
		},
		{
			name:   "getItemOpt",
			source: `Hello {{ user["name"] }}`,
			context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
				Key: value.KeyRefFromString("user"),
				Value: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
					Key:   value.KeyRefFromValue(value.FromString("name")),
					Value: value.FromString("John"),
				}})),
			}})),
			want: "Hello John",
		},
		{
			name:    "sliceString",
			source:  `Hello {{ "Johnson"[:4] }}`,
			context: value.None,
			want:    "Hello John",
		},
		{
			name:    "sliceSeq",
			source:  `Hello {{ ["John", "Paul"][1] }}`,
			context: value.None,
			want:    "Hello Paul",
		},
		{
			name:   "sliceVarElem",
			source: `Hello {{ ["John", name][1] }}`,
			context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
				Key: value.KeyRefFromString("name"), Value: value.FromString("Paul"),
			}})),
			want: "Hello Paul",
		},
		{
			name:    "mapGetItem",
			source:  `Hello {{ {"name": "John"}["name"] }}`,
			context: value.None,
			want:    "Hello John",
		},
		{
			name:   "mapVarValue",
			source: `Hello {{ {"name": name}["name"] }}`,
			context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
				Key: value.KeyRefFromString("name"), Value: value.FromString("Paul"),
			}})),
			want: "Hello Paul",
		},
		{
			name:    "sliceSeqNegativeIndex",
			source:  `Hello {{ ["John", "Paul", "George", "Ringo"][-1] }}`,
			context: value.None,
			want:    "Hello Ringo",
		},
		{
			name:    "addExprString",
			source:  `Hello {{ {"name": "John"}["nam" + "e"] }}`,
			context: value.None,
			want:    "Hello John",
		},
		{
			name:    "addExprInt",
			source:  `Hello {{ ["John", "Paul", "George", "Ringo"][1 + 2] }}`,
			context: value.None,
			want:    "Hello Ringo",
		},
		{
			name:    "addExprFloat",
			source:  `Hello {{ ["John", "Paul", "George", "Ringo"][1.0 + 2.0] }}`,
			context: value.None,
			want:    "Hello Ringo",
		},
		{
			name:    "subExprInt",
			source:  `Hello {{ ["John", "Paul", "George", "Ringo"][3 - 2] }}`,
			context: value.None,
			want:    "Hello Paul",
		},
		{
			name:    "subExprFloat",
			source:  `Hello {{ ["John", "Paul", "George", "Ringo"][3.0 - 2.0] }}`,
			context: value.None,
			want:    "Hello Paul",
		},
		{
			name:   "stringConcat",
			source: `{{ "Hello " ~ name ~ "!" }}`,
			context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
				Key: value.KeyRefFromString("name"), Value: value.FromString("John"),
			}})),
			want: "Hello John!",
		},
		{
			name:    "pow",
			source:  `{{ 2 ** 3 }}`,
			context: value.None,
			want:    "8",
		},
		{
			name:   "ifStmtNoElse",
			source: `{% if down %}I'm down{% endif %}`,
			context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
				Key: value.KeyRefFromString("down"), Value: value.FromBool(true),
			}})),
			want: "I'm down",
		},
		{
			name:   "ifStmtWithElse",
			source: `{% if down %}I'm down{% else %}I'm up{% endif %}`,
			context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
				Key: value.KeyRefFromString("down"), Value: value.FromBool(false),
			}})),
			want: "I'm up",
		},
		{
			name:   "ifExprNoElse",
			source: `{{ "I'm down" if down }}`,
			context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
				Key: value.KeyRefFromString("down"), Value: value.FromBool(true),
			}})),
			want: "I'm down",
		},
		{
			name:   "ifExprWithElse",
			source: `{{ "I'm down" if down else "I'm up" }}`,
			context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
				Key: value.KeyRefFromString("down"), Value: value.FromBool(false),
			}})),
			want: "I'm up",
		},
		{
			name:   "forStmtNoElse",
			source: `{% for name in names %}{{ name }} {% endfor %}`,
			context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
				Key: value.KeyRefFromString("names"), Value: value.FromSlice([]value.Value{
					value.FromString("John"),
					value.FromString("Paul"),
				}),
			}})),
			want: "John Paul ",
		},
		{
			name:    "rawStmt",
			source:  `{% raw %}Hello {{ name }}{% endraw %}`,
			context: value.None,
			want:    "Hello {{ name }}",
		},
		{
			name:    "withStmt",
			source:  `{% with foo = 42 %}{{ foo }}{% endwith %}`,
			context: value.None,
			want:    "42",
		},
		{
			name:    "setStmt",
			source:  `{% set name = "John" %}Hello {{ name }}`,
			context: value.None,
			want:    "Hello John",
		},
		{
			name: "setBlockStmt",
			source: "{% set navigation %}\n" +
				"<li><a href=\"/\">Index</a>\n" +
				"<li><a href=\"/downloads\">Downloads</a>\n" +
				"{% endset %}\n" +
				"<ul>\n" +
				"{{ navigation }}\n" +
				"</ul>\n",
			context: value.None,
			want: "\n" +
				"<ul>\n" +
				"\n" +
				"<li><a href=\"/\">Index</a>\n" +
				"<li><a href=\"/downloads\">Downloads</a>\n" +
				"\n" +
				"</ul>",
		},
		{
			name:    "testIsNotDefined",
			source:  `{% if seq is not defined %}I'm fallback{% endif %}`,
			context: value.None,
			want:    "I'm fallback",
		},
	}
	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := NewEnvironment()
			const templateName = "foo.js"
			err := env.AddTemplate(templateName, tc.source)
			if err != nil {
				t.Fatal(err)
			}
			tpl, err := env.GetTemplate(templateName)
			if err != nil {
				t.Fatal(err)
			}
			got, err := tpl.render(tc.context)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("result mismatch, i=%d, source=%s,\n got=%q,\nwant=%q", i, tc.source, got, tc.want)
			}
		})
	}
}
