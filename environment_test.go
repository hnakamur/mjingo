package mjingo

import (
	"testing"

	"github.com/hnakamur/mjingo/internal"
)

func TestEnvironment(t *testing.T) {
	type testCase struct {
		name    string
		source  string
		context any
		want    string
	}

	runTests := func(t *testing.T, testCases []testCase) {
		for i, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				env := NewEnvironment()
				const templateName = "test.html"
				err := env.AddTemplate(templateName, tc.source)
				if err != nil {
					t.Fatal(err)
				}
				tpl, err := env.GetTemplate(templateName)
				if err != nil {
					t.Fatal(err)
				}
				got, err := tpl.Render(tc.context)
				if err != nil {
					t.Fatal(err)
				}
				if got != tc.want {
					t.Errorf("result mismatch, i=%d, source=%s,\n got=%q,\nwant=%q", i, tc.source, got, tc.want)
				}
			})
		}
	}

	t.Run("expression", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:   "var",
				source: "Hello {{ name }}",
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("name"), Value: internal.FromString("World"),
				}})),
				want: "Hello World",
			},
			{
				name:    "stringExpr",
				source:  `Hello {{ "world" }}`,
				context: internal.None,
				want:    "Hello world",
			},
			{
				name:    "i64Expr",
				source:  `Hello {{ 3 }}`,
				context: internal.None,
				want:    "Hello 3",
			},
			{
				name:    "f64Expr",
				source:  `Hello {{ 3.14 }}`,
				context: internal.None,
				want:    "Hello 3.14",
			},
			{
				name:    "boolExprTrue",
				source:  `Hello {{ true }}`,
				context: internal.None,
				want:    "Hello true",
			},
			{
				name:    "boolExprFalse",
				source:  `Hello {{ False }}`,
				context: internal.None,
				want:    "Hello false",
			},
			{
				name:    "noneExpr",
				source:  `Hello {{ none }}`,
				context: internal.None,
				want:    "Hello none",
			},
			{
				name:   "getFastAttr",
				source: `Hello {{ user.name }}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("user"),
					Value: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
						Key:   internal.KeyRefFromString("name"),
						Value: internal.FromString("John"),
					}})),
				}})),
				want: "Hello John",
			},
			{
				name:   "getItemOpt",
				source: `Hello {{ user["name"] }}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("user"),
					Value: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
						Key:   internal.KeyRefFromValue(internal.FromString("name")),
						Value: internal.FromString("John"),
					}})),
				}})),
				want: "Hello John",
			},
			{
				name:    "sliceString",
				source:  `Hello {{ "Johnson"[:4] }}`,
				context: internal.None,
				want:    "Hello John",
			},
			{
				name:    "sliceSeq",
				source:  `Hello {{ ["John", "Paul"][1] }}`,
				context: internal.None,
				want:    "Hello Paul",
			},
			{
				name:   "sliceVarElem",
				source: `Hello {{ ["John", name][1] }}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("name"), Value: internal.FromString("Paul"),
				}})),
				want: "Hello Paul",
			},
			{
				name:    "mapGetItem",
				source:  `Hello {{ {"name": "John"}["name"] }}`,
				context: internal.None,
				want:    "Hello John",
			},
			{
				name:   "mapVarValue",
				source: `Hello {{ {"name": name}["name"] }}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("name"), Value: internal.FromString("Paul"),
				}})),
				want: "Hello Paul",
			},
			{
				name:    "sliceSeqNegativeIndex",
				source:  `Hello {{ ["John", "Paul", "George", "Ringo"][-1] }}`,
				context: internal.None,
				want:    "Hello Ringo",
			},
			{
				name:    "addExprString",
				source:  `Hello {{ {"name": "John"}["nam" + "e"] }}`,
				context: internal.None,
				want:    "Hello John",
			},
			{
				name:    "addExprInt",
				source:  `Hello {{ ["John", "Paul", "George", "Ringo"][1 + 2] }}`,
				context: internal.None,
				want:    "Hello Ringo",
			},
			{
				name:    "addExprFloat",
				source:  `Hello {{ ["John", "Paul", "George", "Ringo"][1.0 + 2.0] }}`,
				context: internal.None,
				want:    "Hello Ringo",
			},
			{
				name:    "subExprInt",
				source:  `Hello {{ ["John", "Paul", "George", "Ringo"][3 - 2] }}`,
				context: internal.None,
				want:    "Hello Paul",
			},
			{
				name:    "subExprFloat",
				source:  `Hello {{ ["John", "Paul", "George", "Ringo"][3.0 - 2.0] }}`,
				context: internal.None,
				want:    "Hello Paul",
			},
			{
				name:   "stringConcat",
				source: `{{ "Hello " ~ name ~ "!" }}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("name"), Value: internal.FromString("John"),
				}})),
				want: "Hello John!",
			},
			{name: "pow", source: `{{ 2 ** 3 }}`, context: internal.None, want: "8"},
			{name: "mul", source: `{{ 2 * 3 }}`, context: internal.None, want: "6"},
			{name: "div", source: `{{ 3 / 2 }}`, context: internal.None, want: "1.5"},
			{name: "intdiv", source: `{{ 3 // 2 }}`, context: internal.None, want: "1"},
			{name: "rem", source: `{{ 3 % 2 }}`, context: internal.None, want: "1"},
			{name: "neg", source: `{{ -3 }}`, context: internal.None, want: "-3"},
			{name: "notTrue", source: `{{ not 0 }}`, context: internal.None, want: "true"},
			{name: "notFalse", source: `{{ not 1 }}`, context: internal.None, want: "false"},
			{name: "eq", source: `{{ 1 == 1 }}`, context: internal.None, want: "true"},
			{name: "lt", source: `{{ 1 < 2 }}`, context: internal.None, want: "true"},
			{name: "lte", source: `{{ 1 <= 1 }}`, context: internal.None, want: "true"},
			{name: "gt", source: `{{ 2 > 1 }}`, context: internal.None, want: "true"},
			{name: "gte", source: `{{ 1 >= 1 }}`, context: internal.None, want: "true"},
			{name: "inTrue", source: `{{ 1 in [1] }}`, context: internal.None, want: "true"},
			{name: "inFalse", source: `{{ 1 in [2] }}`, context: internal.None, want: "false"},
			{name: "inNot", source: `{{ 1 not in [2] }}`, context: internal.None, want: "true"},
			{name: "tuipleTreatedAsSeq0", source: `{{ () }}`, context: internal.None, want: "[]"},
			{name: "tuipleTreatedAsSeq1", source: `{{ (1,) }}`, context: internal.None, want: "[1]"},
			{name: "tuipleTreatedAsSeq2", source: `{{ (1, 2) }}`, context: internal.None, want: "[1, 2]"},
		})
	})
	t.Run("statement", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:   "ifStmtNoElse",
				source: `{% if down %}I am down{% endif %}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("down"), Value: internal.FromBool(true),
				}})),
				want: "I am down",
			},
			{
				name:   "ifStmtWithElse",
				source: `{% if down %}I am down{% else %}I am up{% endif %}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("down"), Value: internal.FromBool(false),
				}})),
				want: "I am up",
			},
			{
				name:   "ifExprNoElse",
				source: `{{ "I am down" if down }}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("down"), Value: internal.FromBool(true),
				}})),
				want: "I am down",
			},
			{
				name:   "ifExprWithElse",
				source: `{{ "I am down" if down else "I am up" }}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("down"), Value: internal.FromBool(false),
				}})),
				want: "I am up",
			},
			{
				name:   "forStmtNoElse",
				source: `{% for name in names %}{{ name }} {% endfor %}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("names"), Value: internal.FromSlice([]internal.Value{
						internal.FromString("John"),
						internal.FromString("Paul"),
					}),
				}})),
				want: "John Paul ",
			},
			{
				name:    "rawStmt",
				source:  `{% raw %}Hello {{ name }}{% endraw %}`,
				context: internal.None,
				want:    "Hello {{ name }}",
			},
			{
				name:    "withStmt",
				source:  `{% with foo = 42 %}{{ foo }}{% endwith %}`,
				context: internal.None,
				want:    "42",
			},
			{
				name:    "setStmt",
				source:  `{% set name = "John" %}Hello {{ name }}`,
				context: internal.None,
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
				context: internal.None,
				want: "\n" +
					"<ul>\n" +
					"\n" +
					"<li><a href=\"/\">Index</a>\n" +
					"<li><a href=\"/downloads\">Downloads</a>\n" +
					"\n" +
					"</ul>",
			},
			{
				name:   "autoEscapeStmt",
				source: `{% autoescape "html" %}{{ unsafe }}{% endautoescape %} {% autoescape "none" %}{{ unsafe }}{% endautoescape %}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("unsafe"), Value: internal.FromString("<foo>"),
				}})),
				want: "&lt;foo&gt; <foo>",
			},
			{
				name:    "filterStmt",
				source:  `{% filter upper %}hello{% endfilter %} world`,
				context: internal.None,
				want:    "HELLO world",
			},
			// {
			// 	name: "doStmt",
			// 	source: "{% macro dialog(title) %}\n" +
			// 		"Dialog is {{ title }}\n" +
			// 		"{% endmacro %}\n" +
			// 		"{% do dialog(title=\"Hello World\") %}",
			// 	context: internal.None,
			// 	want:    "",
			// },
			// {
			// 	name: "callStmt",
			// 	source: "{% macro dialog(title) %}\n" +
			// 		"Dialog is {{ title }}\n" +
			// 		"{% endmacro %}\n" +
			// 		"{% do dialog(title=\"Hello World\") %}",
			// 	context: internal.None,
			// 	want:    "Dialog is Hello World",
			// },
		})
	})
	t.Run("filter", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:   "escape",
				source: `{{ v|escape }}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("v"), Value: internal.FromString("<br/>"),
				}})),
				want: "&lt;br/&gt;",
			},
			{
				name:   "safeEscape",
				source: `{{ v|safe|e }}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("v"), Value: internal.FromString("<br/>"),
				}})),
				want: "<br/>",
			},
			{name: "lower", source: `{{ "HELLO"|lower }}`, context: internal.None, want: "hello"},
			{name: "upper", source: `{{ "hello"|upper }}`, context: internal.None, want: "HELLO"},
			{name: "title", source: `{{ "hello world"|title }}`, context: internal.None, want: "HELLO WORLD"},
			{name: "capitalize", source: `{{ "hello World"|capitalize }}`, context: internal.None, want: "Hello world"},
			{name: "replace", source: `{{ "Hello World"|replace("Hello", "Goodbye") }}`, context: internal.None, want: "Goodbye World"},
			{name: "count", source: `{{ ["foo", "bar"]|length }}`, context: internal.None, want: "2"},
			{name: "count", source: `{{ "あいう"|length }}`, context: internal.None, want: "3"},
		})
	})
	t.Run("test", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:   "isDefined",
				source: `{% if v is defined %}I am defined{% else %}I am fallback{% endif %}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("v"), Value: internal.None,
				}})),
				want: "I am defined",
			},
			{
				name:    "isNotDefined",
				source:  `{% if v is not defined %}I am fallback{% endif %}`,
				context: internal.None,
				want:    "I am fallback",
			},
			{
				name:   "isNone",
				source: `{% if v is none %}I am none{% else %}I am not none{% endif %}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("v"), Value: internal.None,
				}})),
				want: "I am none",
			},
			{
				name:   "isSafeTrue",
				source: `{% if v is safe %}I am safe{% else %}I am not safe{% endif %}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("v"), Value: internal.FromSafeString("s"),
				}})),
				want: "I am safe",
			},
			{
				name:   "isSafeFalse",
				source: `{% if v is safe %}I am safe{% else %}I am not safe{% endif %}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("v"), Value: internal.FromString("s"),
				}})),
				want: "I am not safe",
			},
			{
				name:   "isEscaped",
				source: `{% if v is escaped %}I am safe{% else %}I am not safe{% endif %}`,
				context: internal.FromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("v"), Value: internal.FromSafeString("s"),
				}})),
				want: "I am safe",
			},
			{name: "isOddTrue", source: `{{ 41 is odd }}`, context: internal.None, want: "true"},
			{name: "isOddValueFalse", source: `{{ 42 is odd }}`, context: internal.None, want: "false"},
			{name: "isOddTypeFalse", source: `{{ "s" is odd }}`, context: internal.None, want: "false"},
			{name: "isEvenTrue", source: `{{ 41 is even }}`, context: internal.None, want: "false"},
			{name: "isEvenValueFalse", source: `{{ 42 is even }}`, context: internal.None, want: "true"},
			{name: "isEvenTypeFalse", source: `{{ "s" is even }}`, context: internal.None, want: "false"},
			{name: "isNumberTrue", source: `{{ 42 is number }}`, context: internal.None, want: "true"},
			{name: "isNumberFalse", source: `{{ "42" is number }}`, context: internal.None, want: "false"},
			{name: "isStringTrue", source: `{{ "42" is string }}`, context: internal.None, want: "true"},
			{name: "isStringFalse", source: `{{ 42 is string }}`, context: internal.None, want: "false"},
			{name: "isSequenceTrue", source: `{{ [1, 2, 3] is sequence }}`, context: internal.None, want: "true"},
			{name: "isSequenceFalse", source: `{{ 42 is sequence }}`, context: internal.None, want: "false"},
			{name: "isMappingTrue", source: `{{ {"foo": "bar"} is mapping }}`, context: internal.None, want: "true"},
			{name: "isMappingFalse", source: `{{ [1, 2, 3] is mapping }}`, context: internal.None, want: "false"},
			{name: "isStartingWithTrue", source: `{{ "foobar" is startingwith("foo") }}`, context: internal.None, want: "true"},
			{name: "isStartingWithFalse", source: `{{ "foobar" is startingwith("bar") }}`, context: internal.None, want: "false"},
			{name: "isEndingWithTrue", source: `{{ "foobar" is endingwith("bar") }}`, context: internal.None, want: "true"},
			{name: "isEndingWithFalse", source: `{{ "foobar" is endingwith("foo") }}`, context: internal.None, want: "false"},
		})
	})
}
