package mjingo

import (
	"testing"

	"github.com/hnakamur/mjingo/value"
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
		})
	})
	t.Run("statement", func(t *testing.T) {
		runTests(t, []testCase{
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
		})
	})
	t.Run("filter", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:   "escape",
				source: `{{ v|escape }}`,
				context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
					Key: value.KeyRefFromString("v"), Value: value.FromString("<br/>"),
				}})),
				want: "&lt;br/&gt;",
			},
			{
				name:   "safeEscape",
				source: `{{ v|safe|e }}`,
				context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
					Key: value.KeyRefFromString("v"), Value: value.FromString("<br/>"),
				}})),
				want: "<br/>",
			},
			{name: "lower", source: `{{ "HELLO"|lower }}`, context: value.None, want: "hello"},
			{name: "upper", source: `{{ "hello"|upper }}`, context: value.None, want: "HELLO"},
			{name: "title", source: `{{ "hello world"|title }}`, context: value.None, want: "HELLO WORLD"},
			{name: "capitalize", source: `{{ "hello World"|capitalize }}`, context: value.None, want: "Hello world"},
		})
	})
	t.Run("test", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:   "isDefined",
				source: `{% if v is defined %}I'm defined{% else %}I'm fallback{% endif %}`,
				context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
					Key: value.KeyRefFromString("v"), Value: value.None,
				}})),
				want: "I'm defined",
			},
			{
				name:    "isNotDefined",
				source:  `{% if v is not defined %}I'm fallback{% endif %}`,
				context: value.None,
				want:    "I'm fallback",
			},
			{
				name:   "isNone",
				source: `{% if v is none %}I'm none{% else %}I'm not none{% endif %}`,
				context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
					Key: value.KeyRefFromString("v"), Value: value.None,
				}})),
				want: "I'm none",
			},
			{
				name:   "isSafeTrue",
				source: `{% if v is safe %}I'm safe{% else %}I'm not safe{% endif %}`,
				context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
					Key: value.KeyRefFromString("v"), Value: value.FromSafeString("s"),
				}})),
				want: "I'm safe",
			},
			{
				name:   "isSafeFalse",
				source: `{% if v is safe %}I'm safe{% else %}I'm not safe{% endif %}`,
				context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
					Key: value.KeyRefFromString("v"), Value: value.FromString("s"),
				}})),
				want: "I'm not safe",
			},
			{
				name:   "isEscaped",
				source: `{% if v is escaped %}I'm safe{% else %}I'm not safe{% endif %}`,
				context: value.FromIndexMap(value.NewIndexMapFromEntries([]value.IndexMapEntry{{
					Key: value.KeyRefFromString("v"), Value: value.FromSafeString("s"),
				}})),
				want: "I'm safe",
			},
			{name: "isOddTrue", source: `{{ 41 is odd }}`, context: value.None, want: "true"},
			{name: "isOddValueFalse", source: `{{ 42 is odd }}`, context: value.None, want: "false"},
			{name: "isOddTypeFalse", source: `{{ "s" is odd }}`, context: value.None, want: "false"},
			{name: "isEvenTrue", source: `{{ 41 is even }}`, context: value.None, want: "false"},
			{name: "isEvenValueFalse", source: `{{ 42 is even }}`, context: value.None, want: "true"},
			{name: "isEvenTypeFalse", source: `{{ "s" is even }}`, context: value.None, want: "false"},
			{name: "isNumberTrue", source: `{{ 42 is number }}`, context: value.None, want: "true"},
			{name: "isNumberFalse", source: `{{ "42" is number }}`, context: value.None, want: "false"},
			{name: "isStringTrue", source: `{{ "42" is string }}`, context: value.None, want: "true"},
			{name: "isStringFalse", source: `{{ 42 is string }}`, context: value.None, want: "false"},
			{name: "isSequenceTrue", source: `{{ [1, 2, 3] is sequence }}`, context: value.None, want: "true"},
			{name: "isSequenceFalse", source: `{{ 42 is sequence }}`, context: value.None, want: "false"},
			{name: "isMappingTrue", source: `{{ {"foo": "bar"} is mapping }}`, context: value.None, want: "true"},
			{name: "isMappingFalse", source: `{{ [1, 2, 3] is mapping }}`, context: value.None, want: "false"},
			{name: "isStartingWithTrue", source: `{{ "foobar" is startingwith("foo") }}`, context: value.None, want: "true"},
			{name: "isStartingWithFalse", source: `{{ "foobar" is startingwith("bar") }}`, context: value.None, want: "false"},
			{name: "isEndingWithTrue", source: `{{ "foobar" is endingwith("bar") }}`, context: value.None, want: "true"},
			{name: "isEndingWithFalse", source: `{{ "foobar" is endingwith("foo") }}`, context: value.None, want: "false"},
		})
	})
}
