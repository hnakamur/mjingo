package mjingo

import (
	"testing"

	"github.com/hnakamur/mjingo/internal/value"
)

func TestSingleTemplate(t *testing.T) {
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
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("name"), Value: value.ValueFromString("World"),
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
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("user"),
					Value: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
						Key:   value.KeyRefFromString("name"),
						Value: value.ValueFromString("John"),
					}})),
				}})),
				want: "Hello John",
			},
			{
				name:   "getItemOpt",
				source: `Hello {{ user["name"] }}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("user"),
					Value: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
						Key:   value.KeyRefFromValue(value.ValueFromString("name")),
						Value: value.ValueFromString("John"),
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
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("name"), Value: value.ValueFromString("Paul"),
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
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("name"), Value: value.ValueFromString("Paul"),
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
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("name"), Value: value.ValueFromString("John"),
				}})),
				want: "Hello John!",
			},
			{name: "pow", source: `{{ 2 ** 3 }}`, context: value.None, want: "8"},
			{name: "mul", source: `{{ 2 * 3 }}`, context: value.None, want: "6"},
			{name: "div", source: `{{ 3 / 2 }}`, context: value.None, want: "1.5"},
			{name: "intdiv", source: `{{ 3 // 2 }}`, context: value.None, want: "1"},
			{name: "rem", source: `{{ 3 % 2 }}`, context: value.None, want: "1"},
			{name: "neg", source: `{{ -3 }}`, context: value.None, want: "-3"},
			{name: "notTrue", source: `{{ not 0 }}`, context: value.None, want: "true"},
			{name: "notFalse", source: `{{ not 1 }}`, context: value.None, want: "false"},
			{name: "eq", source: `{{ 1 == 1 }}`, context: value.None, want: "true"},
			{name: "lt", source: `{{ 1 < 2 }}`, context: value.None, want: "true"},
			{name: "lte", source: `{{ 1 <= 1 }}`, context: value.None, want: "true"},
			{name: "gt", source: `{{ 2 > 1 }}`, context: value.None, want: "true"},
			{name: "gte", source: `{{ 1 >= 1 }}`, context: value.None, want: "true"},
			{name: "inTrue", source: `{{ 1 in [1] }}`, context: value.None, want: "true"},
			{name: "inFalse", source: `{{ 1 in [2] }}`, context: value.None, want: "false"},
			{name: "inNot", source: `{{ 1 not in [2] }}`, context: value.None, want: "true"},
			{name: "tuipleTreatedAsSeq0", source: `{{ () }}`, context: value.None, want: "[]"},
			{name: "tuipleTreatedAsSeq1", source: `{{ (1,) }}`, context: value.None, want: "[1]"},
			{name: "tuipleTreatedAsSeq2", source: `{{ (1, 2) }}`, context: value.None, want: "[1, 2]"},
			{name: "scAnd1", source: `{{ false and false }}`, context: value.None, want: "false"},
			{name: "scAnd2", source: `{{ true and false }}`, context: value.None, want: "false"},
			{name: "scAnd3", source: `{{ true and true }}`, context: value.None, want: "true"},
			{name: "scOr1", source: `{{ false or false }}`, context: value.None, want: "false"},
			{name: "scOr2", source: `{{ false or true }}`, context: value.None, want: "true"},
			{name: "scOr3", source: `{{ true or false }}`, context: value.None, want: "true"},
		})
	})
	t.Run("statement", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:   "ifStmtNoElse",
				source: `{% if down %}I am down{% endif %}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("down"), Value: value.ValueFromBool(true),
				}})),
				want: "I am down",
			},
			{
				name:   "ifStmtWithElse",
				source: `{% if down %}I am down{% else %}I am up{% endif %}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("down"), Value: value.ValueFromBool(false),
				}})),
				want: "I am up",
			},
			{
				name:   "ifExprNoElse",
				source: `{{ "I am down" if down }}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("down"), Value: value.ValueFromBool(true),
				}})),
				want: "I am down",
			},
			{
				name:   "ifExprWithElse",
				source: `{{ "I am down" if down else "I am up" }}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("down"), Value: value.ValueFromBool(false),
				}})),
				want: "I am up",
			},
			{
				name:   "forStmtNoElse",
				source: `{% for name in names %}{{ name }} {% endfor %}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("names"), Value: value.ValueFromSlice([]value.Value{
						value.ValueFromString("John"),
						value.ValueFromString("Paul"),
					}),
				}})),
				want: "John Paul ",
			},
			{
				name:   "forStmtWithElseUnused",
				source: `{% for name in names %}{{ name }} {% else %}no users{% endfor %}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("names"), Value: value.ValueFromSlice([]value.Value{
						value.ValueFromString("John"),
						value.ValueFromString("Paul"),
					}),
				}})),
				want: "John Paul ",
			},
			{
				name:    "forStmtWithElseUsed",
				source:  `{% for name in names %}{{ name }} {% else %}no users{% endfor %}`,
				context: value.Undefined,
				want:    "no users",
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
				name:   "autoEscapeStmt",
				source: `{% autoescape "html" %}{{ unsafe }}{% endautoescape %} {% autoescape "none" %}{{ unsafe }}{% endautoescape %}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("unsafe"), Value: value.ValueFromString("<foo>"),
				}})),
				want: "&lt;foo&gt; <foo>",
			},
			{
				name:    "filterStmt",
				source:  `{% filter upper %}hello{% endfilter %} world`,
				context: value.None,
				want:    "HELLO world",
			},
			{
				name: "doStmt",
				source: "{% macro dialog(title) %}\n" +
					"Dialog is {{ title }}\n" +
					"{% endmacro %}\n" +
					"{% do dialog(title=\"Hello World\") %}",
				context: value.None,
				want:    "\n",
			},
			{
				name: "callStmt",
				source: "{% macro dialog(title) %}\n" +
					"  <div class=\"dialog\">\n" +
					"    <h3>{{ title }}</h3>\n" +
					"    <div class=\"contents\">{{ caller() }}</div>\n" +
					"  </div>\n" +
					"{% endmacro %}\n" +
					"{% call dialog(title=\"Hello World\") %}\n" +
					"  This is the dialog body.\n" +
					"{% endcall %}",
				context: value.None,
				want: "\n\n  <div class=\"dialog\">\n" +
					"    <h3>Hello World</h3>\n" +
					"    <div class=\"contents\">\n  This is the dialog body.\n</div>\n" +
					"  </div>\n",
			},
		})
	})
	t.Run("loopVariable", func(t *testing.T) {
		ctx := value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
			Key: value.KeyRefFromString("users"),
			Value: value.ValueFromSlice([]value.Value{
				value.ValueFromString("John"),
				value.ValueFromString("Paul"),
				value.ValueFromString("George"),
				value.ValueFromString("Ringo"),
			}),
		}}))

		recurCtx := value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
			Key: value.KeyRefFromString("menu"),
			Value: value.ValueFromSlice([]value.Value{
				value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key:   value.KeyRefFromString("href"),
					Value: value.ValueFromString("/menu1"),
				}, {
					Key:   value.KeyRefFromString("title"),
					Value: value.ValueFromString("menu1"),
				}, {
					Key: value.KeyRefFromString("children"),
					Value: value.ValueFromSlice([]value.Value{
						value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
							Key: value.KeyRefFromString("menu"),
							Value: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
								Key:   value.KeyRefFromString("href"),
								Value: value.ValueFromString("/submenu1"),
							}, {
								Key:   value.KeyRefFromString("title"),
								Value: value.ValueFromString("submenu1"),
							}, {
								Key: value.KeyRefFromString("children"),
								Value: value.ValueFromSlice([]value.Value{
									value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
										Key: value.KeyRefFromString("menu"),
										Value: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
											Key:   value.KeyRefFromString("href"),
											Value: value.ValueFromString("/submenu1-1"),
										}, {
											Key:   value.KeyRefFromString("title"),
											Value: value.ValueFromString("submenu1-1"),
										}, {
											Key:   value.KeyRefFromString("children"),
											Value: value.ValueFromSlice([]value.Value{}),
										}})),
									}})),
								}),
							}})),
						}})),
						value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
							Key: value.KeyRefFromString("menu"),
							Value: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
								Key:   value.KeyRefFromString("href"),
								Value: value.ValueFromString("/submenu2"),
							}, {
								Key:   value.KeyRefFromString("title"),
								Value: value.ValueFromString("submenu2"),
							}, {
								Key:   value.KeyRefFromString("children"),
								Value: value.ValueFromSlice([]value.Value{}),
							}})),
						}}))}),
				}})),
			}),
		}}))

		changedCtx := value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
			Key: value.KeyRefFromString("entries"),
			Value: value.ValueFromSlice([]value.Value{
				value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key:   value.KeyRefFromString("category"),
					Value: value.ValueFromString("Go"),
				}, {
					Key:   value.KeyRefFromString("message"),
					Value: value.ValueFromString("Forward Compatibility and Toolchain Management in Go 1.21"),
				}})),
				value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key:   value.KeyRefFromString("category"),
					Value: value.ValueFromString("Go"),
				}, {
					Key:   value.KeyRefFromString("message"),
					Value: value.ValueFromString("Structured Logging with slog"),
				}})),
				value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key:   value.KeyRefFromString("category"),
					Value: value.ValueFromString("Rust"),
				}, {
					Key:   value.KeyRefFromString("message"),
					Value: value.ValueFromString("2022 Annual Rust Survey Results"),
				}})),
				value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key:   value.KeyRefFromString("category"),
					Value: value.ValueFromString("Rust"),
				}, {
					Key:   value.KeyRefFromString("message"),
					Value: value.ValueFromString("Announcing Rust 1.72.0"),
				}})),
			}),
		}}))

		runTests(t, []testCase{{
			name:    "index",
			source:  "{% for user in users %}{{ loop.index }} {{ user }}\n{% endfor %}",
			context: ctx,
			want:    "1 John\n2 Paul\n3 George\n4 Ringo\n",
		}, {
			name:    "index0",
			source:  "{% for user in users %}{{ loop.index0 }} {{ user }}\n{% endfor %}",
			context: ctx,
			want:    "0 John\n1 Paul\n2 George\n3 Ringo\n",
		}, {
			name:    "revindex",
			source:  "{% for user in users %}{{ loop.revindex }} {{ user }}\n{% endfor %}",
			context: ctx,
			want:    "4 John\n3 Paul\n2 George\n1 Ringo\n",
		}, {
			name:    "revindex0",
			source:  "{% for user in users %}{{ loop.revindex0 }} {{ user }}\n{% endfor %}",
			context: ctx,
			want:    "3 John\n2 Paul\n1 George\n0 Ringo\n",
		}, {
			name:    "firstAndLast",
			source:  "{% for user in users %}{{ ', ' if not loop.first }}{{ 'and ' if loop.last }}{{ user }}{% endfor %}",
			context: ctx,
			want:    "John, Paul, George, and Ringo",
		}, {
			name: "depth",
			source: "<ul class=\"menu\">\n" +
				"{% for item in menu recursive %}\n" +
				"  <li><a href=\"{{ item.href }}\">{{ item.title }} (depth={{ loop.depth }})</a>\n" +
				"  {% if item.children %}\n" +
				"	<ul class=\"submenu\">{{ loop(item.children) }}</ul>\n" +
				"  {% endif %}</li>\n" +
				"{% endfor %}\n" +
				"</ul>",
			context: recurCtx,
			want: "<ul class=\"menu\">\n\n  <li><a href=\"/menu1\">menu1 (depth=1)</a>\n" +
				"  \n\t<ul class=\"submenu\">\n  <li><a href=\"\"> (depth=2)</a>\n" +
				"  </li>\n\n  <li><a href=\"\"> (depth=2)</a>\n" +
				"  </li>\n</ul>\n  </li>\n\n</ul>",
		}, {
			name: "depth0",
			source: "<ul class=\"menu\">\n" +
				"{% for item in menu recursive %}\n" +
				"  <li><a href=\"{{ item.href }}\">{{ item.title }} (depth={{ loop.depth0 }})</a>\n" +
				"  {% if item.children %}\n" +
				"	<ul class=\"submenu\">{{ loop(item.children) }}</ul>\n" +
				"  {% endif %}</li>\n" +
				"{% endfor %}\n" +
				"</ul>",
			context: recurCtx,
			want: "<ul class=\"menu\">\n\n  <li><a href=\"/menu1\">menu1 (depth=0)</a>\n" +
				"  \n\t<ul class=\"submenu\">\n  <li><a href=\"\"> (depth=1)</a>\n" +
				"  </li>\n\n  <li><a href=\"\"> (depth=1)</a>\n" +
				"  </li>\n</ul>\n  </li>\n\n</ul>",
		}, {
			name: "changed",
			source: "{% for entry in entries %}\n" +
				"{% if loop.changed(entry.category) %}\n" +
				"  <h2>{{ entry.category }}</h2>\n" +
				"{% endif %}\n" +
				"  <p>{{ entry.message }}</p>\n" +
				"{% endfor %}",
			context: changedCtx,
			want: "\n\n  <h2>Go</h2>\n\n  <p>Forward Compatibility and Toolchain Management in Go 1.21</p>\n" +
				"\n\n  <p>Structured Logging with slog</p>\n" +
				"\n\n  <h2>Rust</h2>\n\n  <p>2022 Annual Rust Survey Results</p>\n" +
				"\n\n  <p>Announcing Rust 1.72.0</p>\n",
		}, {
			name:    "cycle",
			source:  "{% for user in users %}{{ loop.cycle('odd', 'even') }} {{ user }}\n{% endfor %}",
			context: ctx,
			want:    "odd John\neven Paul\nodd George\neven Ringo\n",
		}})
	})
	t.Run("macro", func(t *testing.T) {
		runTests(t, []testCase{{
			name: "closure",
			source: "{% autoescape 'none' %}{% set d = \"closure\" -%}\n" +
				"{% macro example(a, b, c=\"default\") %}{{ [a, b, c, d] }}{% endmacro -%}\n" +
				"{{ example(\"Hello\", \"World\") }}{% endautoescape %}\n",
			context: value.Undefined,
			want:    `["Hello", "World", "default", "closure"]`,
		}})
	})
	t.Run("filter", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:   "escape",
				source: `{{ v|escape }}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("v"), Value: value.ValueFromString("<br/>"),
				}})),
				want: "&lt;br/&gt;",
			},
			{
				name:   "safeEscape",
				source: `{{ v|safe|e }}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("v"), Value: value.ValueFromString("<br/>"),
				}})),
				want: "<br/>",
			},
			{name: "lower", source: `{{ "HELLO"|lower }}`, context: value.None, want: "hello"},
			{name: "upper", source: `{{ "hello"|upper }}`, context: value.None, want: "HELLO"},
			{name: "title", source: `{{ "hello world"|title }}`, context: value.None, want: "HELLO WORLD"},
			{name: "capitalize", source: `{{ "hello World"|capitalize }}`, context: value.None, want: "Hello world"},
			{name: "replace", source: `{{ "Hello World"|replace("Hello", "Goodbye") }}`, context: value.None, want: "Goodbye World"},
			{name: "countSlice", source: `{{ ["foo", "bar"]|length }}`, context: value.None, want: "2"},
			{name: "countStr", source: `{{ "あいう"|length }}`, context: value.None, want: "3"},
			{name: "dictsortCase1", source: `{% autoescape 'none' %}{{ {'a': 4, 'c': 3, 'B': 1}|dictsort }}{% endautoescape %}`, context: value.None, want: `[["a", 4], ["B", 1], ["c", 3]]`},
			{name: "dictsortCase2", source: `{% autoescape 'none' %}{{ {'a': 4, 'c': 3, 'b': 1}|dictsort(by="value") }}{% endautoescape %}`, context: value.None, want: `[["b", 1], ["c", 3], ["a", 4]]`},
			{name: "dictsortCase3", source: `{% autoescape 'none' %}{{ {'a': 4, 'c': 3, 'b': 1}|dictsort(by="value", reverse=true) }}{% endautoescape %}`, context: value.None, want: `[["a", 4], ["c", 3], ["b", 1]]`},
			{name: "dictsortCase4", source: `{% autoescape 'none' %}{{ {'a': 4, 'c': 3, 'B': 1}|dictsort(case_sensitive=true) }}{% endautoescape %}`, context: value.None, want: `[["B", 1], ["a", 4], ["c", 3]]`},
			{name: "sortCase1", source: `{% autoescape 'none' %}{{ ['a', 'c', 'B']|sort }}{% endautoescape %}`, context: value.None, want: `["a", "B", "c"]`},
			{name: "sortCase2", source: `{% autoescape 'none' %}{{ ['a', 'c', 'B']|sort(reverse=true) }}{% endautoescape %}`, context: value.None, want: `["c", "B", "a"]`},
			{name: "sortCase3", source: `{% autoescape 'none' %}{{ ['a', 'c', 'B']|sort(case_sensitive=true) }}{% endautoescape %}`, context: value.None, want: `["B", "a", "c"]`},
			{name: "sortCase2", source: `{% autoescape 'none' %}{{ [{"name": "John", "id": 1}, {"name": "Paul", "id": 2}]|sort(attribute="id", reverse=true) }}{% endautoescape %}`, context: value.None, want: `[{"name": "Paul", "id": 2}, {"name": "John", "id": 1}]`},
			{
				name:    "items",
				source:  `{% for key, value in {'a': 1, 'b': 2}|items %}{%if not loop.first %}, {% endif %}{{ key }}: {{ value }}{% endfor %}`,
				context: value.None,
				want:    "a: 1, b: 2",
			},
			{name: "joinStrNoJoiner", source: `{{ "あいう"|join }}`, context: value.None, want: "あいう"},
			{name: "joinStrWithJoiner", source: `{{ "あいう"|join(",") }}`, context: value.None, want: "あ,い,う"},
			{name: "joinSeqNoJoiner", source: `{{ [1, 2, 3]|join }}`, context: value.None, want: "123"},
			{name: "joinSeqWithJoiner", source: `{{ [1, 2, 3]|join(", ") }}`, context: value.None, want: "1, 2, 3"},
			{name: "reverseStr", source: `{{ "あいう"|reverse }}`, context: value.None, want: "ういあ"},
			{name: "reverseSeq", source: `{{ [1, 2, 3]|reverse }}`, context: value.None, want: "[3, 2, 1]"},
			{name: "trimWithCutset", source: `{{ "¡¡¡Hello, Gophers!!!"|trim("!¡") }}`, context: value.None, want: "Hello, Gophers"},
			{name: "trimNoCutset", source: `{{ " \tHello, Gophers\n "|trim }}`, context: value.None, want: "Hello, Gophers"},
			{name: "defaultNoArg", source: `{{ undefined|default }}`, context: value.None, want: ""},
			{name: "defaultStrArg", source: `{{ undefined|default("hello") }}`, context: value.None, want: "hello"},
			{name: "defaultIntArg", source: `{{ undefined|default(2) }}`, context: value.None, want: "2"},
			{name: "defaultAlias", source: `{{ undefined|d(2) }}`, context: value.None, want: "2"},
			{name: "defaultAliasWithVal", source: `{{ 3|d }}`, context: value.None, want: "3"},
			{name: "roundDefPrec", source: `{{ 42.5|round }}`, context: value.None, want: "43.0"},
			{name: "roundWithPrec", source: `{{ 42.45|round(1) }}`, context: value.None, want: "42.5"},
			{name: "absI64", source: `{{ -3|abs }}`, context: value.None, want: "3"},
			{name: "absI128", source: `{{ (-9223372036854775807 * 2)|abs }}`, context: value.None, want: "18446744073709551614"},
			{name: "absF64", source: `{{ -3.2|abs }}`, context: value.None, want: "3.2"},
			{name: "attr", source: `{{ {'a': 1, 'b': 2}|attr('b') }}`, context: value.None, want: "2"},
			{name: "firstStr", source: `{{ "あいう"|first }}`, context: value.None, want: "あ"},
			{name: "firstSeq", source: `{{ [1, 2, 3]|first }}`, context: value.None, want: "1"},
			{name: "lastStr", source: `{{ "あいう"|last }}`, context: value.None, want: "う"},
			{name: "lastSeq", source: `{{ [1, 2, 3]|last }}`, context: value.None, want: "3"},
			{name: "minSeq", source: `{{ [1, 2, 3]|min }}`, context: value.None, want: "1"},
			{name: "maxSeq", source: `{{ [1, 2, 3]|max }}`, context: value.None, want: "3"},
			{name: "listStr", source: `{% autoescape 'none' %}{{ "あいう"|list }}{% endautoescape %}`, context: value.None, want: `["あ", "い", "う"]`},
			{name: "boolTrue", source: `{{ 1|bool }}`, context: value.None, want: "true"},
			{name: "boolFalse", source: `{{ 0|bool }}`, context: value.None, want: "false"},
			{name: "batchNoFiller", source: `{{ [1, 2, 3, 4, 5]|batch(3) }}`, context: value.None, want: "[[1, 2, 3], [4, 5]]"},
			{name: "batchWithFiller", source: `{{ [1, 2, 3, 4, 5]|batch(3, 0) }}`, context: value.None, want: "[[1, 2, 3], [4, 5, 0]]"},
			{name: "sliceNoFiller", source: `{{ [1, 2, 3, 4, 5]|slice(3) }}`, context: value.None, want: "[[1, 2], [3, 4], [5]]"},
			{name: "sliceWithFiller", source: `{{ [1, 2, 3, 4, 5]|slice(3, 0) }}`, context: value.None, want: "[[1, 2], [3, 4], [5, 0]]"},
			{name: "indentCase1", source: `{{ "line1\n  line2\n\n  line3\n"|indent(2) }}`, context: value.None, want: "line1\n    line2\n\n    line3"},
			{name: "indentCase2", source: `{{ "line1\n  line2\n\n  line3\n"|indent(2, true) }}`, context: value.None, want: "  line1\n    line2\n\n    line3"},
			{name: "indentCase3", source: `{{ "line1\n  line2\n\n  line3\n"|indent(2, false, true) }}`, context: value.None, want: "line1\n    line2\n  \n    line3"},
			{name: "selectCase1", source: `{{ [1, 2, 3, 4, 5]|select("odd") }}`, context: value.None, want: "[1, 3, 5]"},
			{name: "rejectCase1", source: `{{ [1, 2, 3, 4, 5]|reject("odd") }}`, context: value.None, want: "[2, 4]"},
			{name: "selectattrCase1", source: `{% autoescape 'none' %}{{ [{"name": "John", "is_active": false}, {"name": "Paul", "is_active": true}]|selectattr("is_active") }}{% endautoescape %}`, context: value.None, want: `[{"name": "Paul", "is_active": true}]`},
			{name: "selectattrCase2", source: `{% autoescape 'none' %}{{ [{"name": "John", "id": 1}, {"name": "Paul", "id": 2}]|selectattr("id", "even") }}{% endautoescape %}`, context: value.None, want: `[{"name": "Paul", "id": 2}]`},
			{name: "rejectattrCase1", source: `{% autoescape 'none' %}{{ [{"name": "John", "is_active": false}, {"name": "Paul", "is_active": true}]|rejectattr("is_active") }}{% endautoescape %}`, context: value.None, want: `[{"name": "John", "is_active": false}]`},
			{name: "rejectattrCase2", source: `{% autoescape 'none' %}{{ [{"name": "John", "id": 1}, {"name": "Paul", "id": 2}]|rejectattr("id", "even") }}{% endautoescape %}`, context: value.None, want: `[{"name": "John", "id": 1}]`},
			{name: "unique", source: `{% autoescape 'none' %}{{ ['foo', 'bar', 'foobar', 'foobar']|unique }}{% endautoescape %}`, context: value.None, want: `["foo", "bar", "foobar"]`},
			{name: "mapCase1", source: `{% autoescape 'none' %}{{ [{"name": "John", "id": 1}, {"name": "Paul", "id": 2}]|map(attribute="name")|join(', ') }}{% endautoescape %}`, context: value.None, want: `John, Paul`},
			{name: "mapCase2", source: `{% autoescape 'none' %}{{ [-1, -2, 3, 4, -5]|map("abs") }}{% endautoescape %}`, context: value.None, want: `[1, 2, 3, 4, 5]`},
		})
	})
	t.Run("test", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:   "isDefined",
				source: `{% if v is defined %}I am defined{% else %}I am fallback{% endif %}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("v"), Value: value.None,
				}})),
				want: "I am defined",
			},
			{
				name:    "isNotDefined",
				source:  `{% if v is not defined %}I am fallback{% endif %}`,
				context: value.None,
				want:    "I am fallback",
			},
			{
				name:   "isNone",
				source: `{% if v is none %}I am none{% else %}I am not none{% endif %}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("v"), Value: value.None,
				}})),
				want: "I am none",
			},
			{
				name:   "isSafeTrue",
				source: `{% if v is safe %}I am safe{% else %}I am not safe{% endif %}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("v"), Value: value.ValueFromSafeString("s"),
				}})),
				want: "I am safe",
			},
			{
				name:   "isSafeFalse",
				source: `{% if v is safe %}I am safe{% else %}I am not safe{% endif %}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("v"), Value: value.ValueFromString("s"),
				}})),
				want: "I am not safe",
			},
			{
				name:   "isEscaped",
				source: `{% if v is escaped %}I am safe{% else %}I am not safe{% endif %}`,
				context: value.ValueFromIndexMap(value.ValueMapFromEntries([]value.ValueMapEntry{{
					Key: value.KeyRefFromString("v"), Value: value.ValueFromSafeString("s"),
				}})),
				want: "I am safe",
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
			{name: "eqTrue", source: `{{ 41 is eq(41) }}`, context: value.None, want: "true"},
			{name: "eqFalse", source: `{{ 41 is eq(42) }}`, context: value.None, want: "false"},
			{name: "equaltoTrue", source: `{{ 41 is equalto(41) }}`, context: value.None, want: "true"},
			{name: "equaltoFalse", source: `{{ 41 is equalto(42) }}`, context: value.None, want: "false"},
			{name: "==True", source: `{{ 41 == 41 }}`, context: value.None, want: "true"},
			{name: "==False", source: `{{ 41 == 42 }}`, context: value.None, want: "false"},
			{name: "neTrue", source: `{{ 41 is ne(42) }}`, context: value.None, want: "true"},
			{name: "neFalse", source: `{{ 41 is ne(41) }}`, context: value.None, want: "false"},
			{name: "!=True", source: `{{ 41 != 42 }}`, context: value.None, want: "true"},
			{name: "!=False", source: `{{ 41 != 41 }}`, context: value.None, want: "false"},
			{name: "ltTrue", source: `{{ 41 is lt(42) }}`, context: value.None, want: "true"},
			{name: "ltFalse", source: `{{ 41 is lt(41) }}`, context: value.None, want: "false"},
			{name: "lessthanTrue", source: `{{ 41 is lessthan(42) }}`, context: value.None, want: "true"},
			{name: "lessthanFalse", source: `{{ 41 is lessthan(41) }}`, context: value.None, want: "false"},
			{name: "<True", source: `{{ 41 < 42 }}`, context: value.None, want: "true"},
			{name: "<False", source: `{{ 41 < 41 }}`, context: value.None, want: "false"},
			{name: "leTrue", source: `{{ 41 is le(41) }}`, context: value.None, want: "true"},
			{name: "leFalse", source: `{{ 41 is le(40) }}`, context: value.None, want: "false"},
			{name: "<=True", source: `{{ 41 <= 41 }}`, context: value.None, want: "true"},
			{name: "<=False", source: `{{ 41 <= 40 }}`, context: value.None, want: "false"},
			{name: "gtTrue", source: `{{ 42 is gt(41) }}`, context: value.None, want: "true"},
			{name: "gtFalse", source: `{{ 41 is gt(41) }}`, context: value.None, want: "false"},
			{name: "greaterthanTrue", source: `{{ 42 is greaterthan(41) }}`, context: value.None, want: "true"},
			{name: "greaterthanFalse", source: `{{ 41 is greaterthan(41) }}`, context: value.None, want: "false"},
			{name: ">True", source: `{{ 42 > 41 }}`, context: value.None, want: "true"},
			{name: ">False", source: `{{ 41 > 41 }}`, context: value.None, want: "false"},
			{name: "geTrue", source: `{{ 42 is ge(42) }}`, context: value.None, want: "true"},
			{name: "geFalse", source: `{{ 40 is ge(41) }}`, context: value.None, want: "false"},
			{name: ">=True", source: `{{ 41 >= 41 }}`, context: value.None, want: "true"},
			{name: ">=False", source: `{{ 40 >= 41 }}`, context: value.None, want: "false"},
			{name: "isInTrue", source: `{{ 1 is in([1, 2]) }}`, context: value.None, want: "true"},
			{name: "isInFalse", source: `{{ 3 is in([1, 2]) }}`, context: value.None, want: "false"},
			{name: "isTrueTrue", source: `{{ true is true }}`, context: value.None, want: "true"},
			{name: "isTrueFalse", source: `{{ 1 is true }}`, context: value.None, want: "false"},
			{name: "isFalseTrue", source: `{{ false is false }}`, context: value.None, want: "true"},
			{name: "isFalseFalse", source: `{{ 0 is false }}`, context: value.None, want: "false"},
			{name: "isFilterTrue", source: `{{ "escape" is filter }}`, context: value.None, want: "true"},
			{name: "isFilterFalse", source: `{{ "no_such_filter" is filter }}`, context: value.None, want: "false"},
			{name: "isTestTrue", source: `{{ "defined" is test }}`, context: value.None, want: "true"},
			{name: "isTestFalse", source: `{{ "no_such_test" is test }}`, context: value.None, want: "false"},
		})
	})
	t.Run("function", func(t *testing.T) {
		runTests(t, []testCase{
			{name: "rangeJustUpper", source: "{{ range(3) }}", context: value.None, want: "[0, 1, 2]"},
			{name: "rangeLowerUpper", source: "{{ range(2, 4) }}", context: value.None, want: "[2, 3]"},
			{name: "rangeLowerUpperStep", source: "{{ range(2, 9, 3) }}", context: value.None, want: "[2, 5, 8]"},
			{name: "dictEmpty", source: "{{ dict()['foo']|default(1) }}", context: value.None, want: "1"},
			{name: "dictNonEmpty", source: "{{ dict(foo='bar')['foo'] }}", context: value.None, want: "bar"},
		})
	})
}

func TestErrorSingleTemplate(t *testing.T) {
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
				var got string
				if _, err := tpl.Render(tc.context); err != nil {
					got = err.Error()
				}
				if got != tc.want {
					t.Errorf("error mismatch, i=%d, source=%s, got=%s, want=%s", i, tc.source, got, tc.want)
				}
			})
		}
	}
	t.Run("function", func(t *testing.T) {
		runTests(t, []testCase{
			{name: "rangeNoArgErr", source: "{{ range() }}", context: value.None, want: "missing argument"},
			{name: "rangeTooManyArgErr", source: "{{ range(1, 2, 3, 4) }}", context: value.None, want: "too many arguments"},
		})
	})
}

func TestMultiTemplates(t *testing.T) {
	type strTemplate struct {
		name   string
		source string
	}

	type testCase struct {
		name      string
		templates []strTemplate
		context   any
		want      string
	}

	runTests := func(t *testing.T, testCases []testCase) {
		t.Helper()
		for i, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				env := NewEnvironment()
				for _, tpl := range tc.templates {
					err := env.AddTemplate(tpl.name, tpl.source)
					if err != nil {
						t.Fatal(err)
					}
				}
				tpl, err := env.GetTemplate(tc.templates[0].name)
				if err != nil {
					t.Fatal(err)
				}
				got, err := tpl.Render(tc.context)
				if err != nil {
					t.Fatal(err)
				}
				if got != tc.want {
					t.Errorf("result mismatch, i=%d, source=%s,\n got=%q,\nwant=%q", i, tc.templates[0].source, got, tc.want)
				}
			})
		}
	}

	t.Run("include", func(t *testing.T) {
		runTests(t, []testCase{{
			name: "simple",
			templates: []strTemplate{{
				name: "main.html",
				source: "{% include 'header.html' %}\n" +
					"  Body\n" +
					"{% include 'footer.html' %}",
			}, {
				name:   "header.html",
				source: "Header",
			}, {
				name:   "footer.html",
				source: "Footer",
			}},
			context: value.None,
			want:    "Header\n  Body\nFooter",
		}})
	})
	t.Run("extends", func(t *testing.T) {
		runTests(t, []testCase{{
			name: "simple",
			templates: []strTemplate{{
				name: "child.html",
				source: "{% extends \"base.html\" %}\n" +
					"{% block title %}Index{% endblock %}\n" +
					"{% block head %}\n" +
					"  {{ super() }}\n" +
					"  <style type=\"text/css\">\n" +
					"	.important { color: #336699; }\n" +
					"  </style>\n" +
					"{% endblock %}\n" +
					"{% block body %}\n" +
					"  <h1>Index</h1>\n" +
					"  <p class=\"important\">\n" +
					"	Welcome to my awesome homepage.\n" +
					"  </p>\n" +
					"{% endblock body %}",
			}, {
				name: "base.html",
				source: "<!doctype html>\n" +
					"{% block head %}\n" +
					"<title>{% block title %}{% endblock %}</title>\n" +
					"{% endblock %}\n" +
					"{% block body %}{% endblock %}",
			}},
			context: value.None,
			want: "<!doctype html>\n\n  \n" +
				"<title>Index</title>\n\n" +
				"  <style type=\"text/css\">\n" +
				"	.important { color: #336699; }\n" +
				"  </style>\n\n\n" +
				"  <h1>Index</h1>\n" +
				"  <p class=\"important\">\n" +
				"	Welcome to my awesome homepage.\n" +
				"  </p>\n",
		}})
	})
	t.Run("import", func(t *testing.T) {
		runTests(t, []testCase{{
			name: "simple",
			templates: []strTemplate{{
				name: "main.html",
				source: "{% import \"my_template.html\" as helpers %}\n" +
					"{{ helpers.my_macro(helpers.my_variable) }}",
			}, {
				name: "my_template.html",
				source: "{% macro my_macro(name) %}\n" +
					"Hello {{ name }}\n" +
					"{% endmacro %}\n" +
					"{% set my_variable = \"World\" %}\n",
			}},
			context: value.None,
			want:    "\n\nHello World\n",
		}})
	})
	t.Run("from", func(t *testing.T) {
		runTests(t, []testCase{{
			name: "simple",
			templates: []strTemplate{{
				name: "main.html",
				source: "{% from \"my_template.html\" import my_macro, my_alias as alias %}\n" +
					"{{ my_macro(\"World\") }}\n" +
					"{{ alias(\"日本\") }}",
			}, {
				name: "my_template.html",
				source: "{% macro my_macro(name) %}\n" +
					"Hello {{ name }}\n" +
					"{% endmacro %}\n" +
					"{% set my_alias = my_macro %}\n",
			}},
			context: value.None,
			want: "\n\nHello World\n" +
				"\n\nHello 日本\n",
		}})
	})
}

func TestSingleTemplWithGoVal(t *testing.T) {
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
				context := value.ValueFromGoValue(tc.context, value.WithStructTag("json"))
				got, err := tpl.Render(context)
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
		type user struct {
			Name string `json:"name"`
			ID   *int   `json:"id"`
		}
		type users struct {
			Users []user `json:"users"`
		}

		intPtr := func(n int) *int {
			return &n
		}

		runTests(t, []testCase{
			{
				name:   "var",
				source: "Hello {{ name }}",
				context: map[string]any{
					"name": "World",
				},
				want: "Hello World",
			},
			{
				name:   "loopIndex",
				source: "{% for user in users %}{{ loop.index }} {{ user }}\n{% endfor %}",
				context: struct {
					Users []string `json:"users"`
				}{
					Users: []string{"John", "Paul", "George", "Ringo"},
				},
				want: "1 John\n2 Paul\n3 George\n4 Ringo\n",
			},
			{
				name:   "selectattrCase2",
				source: `{% autoescape 'none' %}{{ [{"name": "John", "id": 1}, {"name": "Paul", "id": 2}]|selectattr("id", "even") }}{% endautoescape %}`,
				context: users{
					Users: []user{
						{Name: "John", ID: intPtr(1)},
						{Name: "Paul", ID: intPtr(2)},
					},
				},
				want: `[{"name": "Paul", "id": 2}]`,
			},
		})
	})
}
