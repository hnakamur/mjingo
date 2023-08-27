package mjingo

import (
	"testing"

	"github.com/hnakamur/mjingo/internal"
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
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("name"), Value: internal.ValueFromString("World"),
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
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("user"),
					Value: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
						Key:   internal.KeyRefFromString("name"),
						Value: internal.ValueFromString("John"),
					}})),
				}})),
				want: "Hello John",
			},
			{
				name:   "getItemOpt",
				source: `Hello {{ user["name"] }}`,
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("user"),
					Value: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
						Key:   internal.KeyRefFromValue(internal.ValueFromString("name")),
						Value: internal.ValueFromString("John"),
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
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("name"), Value: internal.ValueFromString("Paul"),
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
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("name"), Value: internal.ValueFromString("Paul"),
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
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("name"), Value: internal.ValueFromString("John"),
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
			{name: "scAnd1", source: `{{ false and false }}`, context: internal.None, want: "false"},
			{name: "scAnd2", source: `{{ true and false }}`, context: internal.None, want: "false"},
			{name: "scAnd3", source: `{{ true and true }}`, context: internal.None, want: "true"},
			{name: "scOr1", source: `{{ false or false }}`, context: internal.None, want: "false"},
			{name: "scOr2", source: `{{ false or true }}`, context: internal.None, want: "true"},
			{name: "scOr3", source: `{{ true or false }}`, context: internal.None, want: "true"},
		})
	})
	t.Run("statement", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:   "ifStmtNoElse",
				source: `{% if down %}I am down{% endif %}`,
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("down"), Value: internal.ValueFromBool(true),
				}})),
				want: "I am down",
			},
			{
				name:   "ifStmtWithElse",
				source: `{% if down %}I am down{% else %}I am up{% endif %}`,
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("down"), Value: internal.ValueFromBool(false),
				}})),
				want: "I am up",
			},
			{
				name:   "ifExprNoElse",
				source: `{{ "I am down" if down }}`,
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("down"), Value: internal.ValueFromBool(true),
				}})),
				want: "I am down",
			},
			{
				name:   "ifExprWithElse",
				source: `{{ "I am down" if down else "I am up" }}`,
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("down"), Value: internal.ValueFromBool(false),
				}})),
				want: "I am up",
			},
			{
				name:   "forStmtNoElse",
				source: `{% for name in names %}{{ name }} {% endfor %}`,
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("names"), Value: internal.ValueFromSlice([]internal.Value{
						internal.ValueFromString("John"),
						internal.ValueFromString("Paul"),
					}),
				}})),
				want: "John Paul ",
			},
			{
				name:   "forStmtWithElseUnused",
				source: `{% for name in names %}{{ name }} {% else %}no users{% endfor %}`,
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("names"), Value: internal.ValueFromSlice([]internal.Value{
						internal.ValueFromString("John"),
						internal.ValueFromString("Paul"),
					}),
				}})),
				want: "John Paul ",
			},
			{
				name:    "forStmtWithElseUsed",
				source:  `{% for name in names %}{{ name }} {% else %}no users{% endfor %}`,
				context: internal.Undefined,
				want:    "no users",
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
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("unsafe"), Value: internal.ValueFromString("<foo>"),
				}})),
				want: "&lt;foo&gt; <foo>",
			},
			{
				name:    "filterStmt",
				source:  `{% filter upper %}hello{% endfilter %} world`,
				context: internal.None,
				want:    "HELLO world",
			},
			{
				name: "doStmt",
				source: "{% macro dialog(title) %}\n" +
					"Dialog is {{ title }}\n" +
					"{% endmacro %}\n" +
					"{% do dialog(title=\"Hello World\") %}",
				context: internal.None,
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
				context: internal.None,
				want: "\n\n  <div class=\"dialog\">\n" +
					"    <h3>Hello World</h3>\n" +
					"    <div class=\"contents\">\n  This is the dialog body.\n</div>\n" +
					"  </div>\n",
			},
		})
	})
	t.Run("loopVariable", func(t *testing.T) {
		ctx := internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
			Key: internal.KeyRefFromString("users"),
			Value: internal.ValueFromSlice([]internal.Value{
				internal.ValueFromString("John"),
				internal.ValueFromString("Paul"),
				internal.ValueFromString("George"),
				internal.ValueFromString("Ringo"),
			}),
		}}))

		recurCtx := internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
			Key: internal.KeyRefFromString("menu"),
			Value: internal.ValueFromSlice([]internal.Value{
				internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key:   internal.KeyRefFromString("href"),
					Value: internal.ValueFromString("/menu1"),
				}, {
					Key:   internal.KeyRefFromString("title"),
					Value: internal.ValueFromString("menu1"),
				}, {
					Key: internal.KeyRefFromString("children"),
					Value: internal.ValueFromSlice([]internal.Value{
						internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
							Key: internal.KeyRefFromString("menu"),
							Value: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
								Key:   internal.KeyRefFromString("href"),
								Value: internal.ValueFromString("/submenu1"),
							}, {
								Key:   internal.KeyRefFromString("title"),
								Value: internal.ValueFromString("submenu1"),
							}, {
								Key: internal.KeyRefFromString("children"),
								Value: internal.ValueFromSlice([]internal.Value{
									internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
										Key: internal.KeyRefFromString("menu"),
										Value: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
											Key:   internal.KeyRefFromString("href"),
											Value: internal.ValueFromString("/submenu1-1"),
										}, {
											Key:   internal.KeyRefFromString("title"),
											Value: internal.ValueFromString("submenu1-1"),
										}, {
											Key:   internal.KeyRefFromString("children"),
											Value: internal.ValueFromSlice([]internal.Value{}),
										}})),
									}})),
								}),
							}})),
						}})),
						internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
							Key: internal.KeyRefFromString("menu"),
							Value: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
								Key:   internal.KeyRefFromString("href"),
								Value: internal.ValueFromString("/submenu2"),
							}, {
								Key:   internal.KeyRefFromString("title"),
								Value: internal.ValueFromString("submenu2"),
							}, {
								Key:   internal.KeyRefFromString("children"),
								Value: internal.ValueFromSlice([]internal.Value{}),
							}})),
						}}))}),
				}})),
			}),
		}}))

		changedCtx := internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
			Key: internal.KeyRefFromString("entries"),
			Value: internal.ValueFromSlice([]internal.Value{
				internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key:   internal.KeyRefFromString("category"),
					Value: internal.ValueFromString("Go"),
				}, {
					Key:   internal.KeyRefFromString("message"),
					Value: internal.ValueFromString("Forward Compatibility and Toolchain Management in Go 1.21"),
				}})),
				internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key:   internal.KeyRefFromString("category"),
					Value: internal.ValueFromString("Go"),
				}, {
					Key:   internal.KeyRefFromString("message"),
					Value: internal.ValueFromString("Structured Logging with slog"),
				}})),
				internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key:   internal.KeyRefFromString("category"),
					Value: internal.ValueFromString("Rust"),
				}, {
					Key:   internal.KeyRefFromString("message"),
					Value: internal.ValueFromString("2022 Annual Rust Survey Results"),
				}})),
				internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key:   internal.KeyRefFromString("category"),
					Value: internal.ValueFromString("Rust"),
				}, {
					Key:   internal.KeyRefFromString("message"),
					Value: internal.ValueFromString("Announcing Rust 1.72.0"),
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
			source: "{% set d = \"closure\" -%}\n" +
				"{% macro example(a, b, c=\"default\") %}{{ [a, b, c, d] }}{% endmacro -%}\n" +
				"{{ example(\"Hello\", \"World\") }}\n",
			context: internal.Undefined,
			want:    "[Hello, World, default, closure]",
		}})
	})
	t.Run("filter", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:   "escape",
				source: `{{ v|escape }}`,
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("v"), Value: internal.ValueFromString("<br/>"),
				}})),
				want: "&lt;br/&gt;",
			},
			{
				name:   "safeEscape",
				source: `{{ v|safe|e }}`,
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("v"), Value: internal.ValueFromString("<br/>"),
				}})),
				want: "<br/>",
			},
			{name: "lower", source: `{{ "HELLO"|lower }}`, context: internal.None, want: "hello"},
			{name: "upper", source: `{{ "hello"|upper }}`, context: internal.None, want: "HELLO"},
			{name: "title", source: `{{ "hello world"|title }}`, context: internal.None, want: "HELLO WORLD"},
			{name: "capitalize", source: `{{ "hello World"|capitalize }}`, context: internal.None, want: "Hello world"},
			{name: "replace", source: `{{ "Hello World"|replace("Hello", "Goodbye") }}`, context: internal.None, want: "Goodbye World"},
			{name: "countSlice", source: `{{ ["foo", "bar"]|length }}`, context: internal.None, want: "2"},
			{name: "countStr", source: `{{ "あいう"|length }}`, context: internal.None, want: "3"},
			{
				name:    "items",
				source:  `{% for key, value in {'a': 1, 'b': 2}|items %}{%if not loop.first %}, {% endif %}{{ key }}: {{ value }}{% endfor %}`,
				context: internal.None,
				want:    "a: 1, b: 2",
			},
			{name: "joinStrNoJoiner", source: `{{ "あいう"|join }}`, context: internal.None, want: "あいう"},
			{name: "joinStrWithJoiner", source: `{{ "あいう"|join(",") }}`, context: internal.None, want: "あ,い,う"},
			{name: "joinSeqNoJoiner", source: `{{ [1, 2, 3]|join }}`, context: internal.None, want: "123"},
			{name: "joinSeqWithJoiner", source: `{{ [1, 2, 3]|join(", ") }}`, context: internal.None, want: "1, 2, 3"},
			{name: "reverseStr", source: `{{ "あいう"|reverse }}`, context: internal.None, want: "ういあ"},
			{name: "reverseSeq", source: `{{ [1, 2, 3]|reverse }}`, context: internal.None, want: "[3, 2, 1]"},
			{name: "trimWithCutset", source: `{{ "¡¡¡Hello, Gophers!!!"|trim("!¡") }}`, context: internal.None, want: "Hello, Gophers"},
			{name: "trimNoCutset", source: `{{ " \tHello, Gophers\n "|trim }}`, context: internal.None, want: "Hello, Gophers"},
			{name: "defaultNoArg", source: `{{ undefined|default }}`, context: internal.None, want: ""},
			{name: "defaultStrArg", source: `{{ undefined|default("hello") }}`, context: internal.None, want: "hello"},
			{name: "defaultIntArg", source: `{{ undefined|default(2) }}`, context: internal.None, want: "2"},
			{name: "defaultAlias", source: `{{ undefined|d(2) }}`, context: internal.None, want: "2"},
			{name: "defaultAliasWithVal", source: `{{ 3|d }}`, context: internal.None, want: "3"},
			{name: "roundDefPrec", source: `{{ 42.5|round }}`, context: internal.None, want: "43.0"},
			{name: "roundWithPrec", source: `{{ 42.45|round(1) }}`, context: internal.None, want: "42.5"},
			{name: "absI64", source: `{{ -3|abs }}`, context: internal.None, want: "3"},
			{name: "absI128", source: `{{ (-9223372036854775807 * 2)|abs }}`, context: internal.None, want: "18446744073709551614"},
			{name: "absF64", source: `{{ -3.2|abs }}`, context: internal.None, want: "3.2"},
			{name: "attr", source: `{{ {'a': 1, 'b': 2}|attr('b') }}`, context: internal.None, want: "2"},
			{name: "firstStr", source: `{{ "あいう"|first }}`, context: internal.None, want: "あ"},
			{name: "firstSeq", source: `{{ [1, 2, 3]|first }}`, context: internal.None, want: "1"},
			{name: "lastStr", source: `{{ "あいう"|last }}`, context: internal.None, want: "う"},
			{name: "lastSeq", source: `{{ [1, 2, 3]|last }}`, context: internal.None, want: "3"},
		})
	})
	t.Run("test", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:   "isDefined",
				source: `{% if v is defined %}I am defined{% else %}I am fallback{% endif %}`,
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
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
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("v"), Value: internal.None,
				}})),
				want: "I am none",
			},
			{
				name:   "isSafeTrue",
				source: `{% if v is safe %}I am safe{% else %}I am not safe{% endif %}`,
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("v"), Value: internal.ValueFromSafeString("s"),
				}})),
				want: "I am safe",
			},
			{
				name:   "isSafeFalse",
				source: `{% if v is safe %}I am safe{% else %}I am not safe{% endif %}`,
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("v"), Value: internal.ValueFromString("s"),
				}})),
				want: "I am not safe",
			},
			{
				name:   "isEscaped",
				source: `{% if v is escaped %}I am safe{% else %}I am not safe{% endif %}`,
				context: internal.ValueFromIndexMap(internal.NewIndexMapFromEntries([]internal.IndexMapEntry{{
					Key: internal.KeyRefFromString("v"), Value: internal.ValueFromSafeString("s"),
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
			{name: "eqTrue", source: `{{ 41 is eq(41) }}`, context: internal.None, want: "true"},
			{name: "eqFalse", source: `{{ 41 is eq(42) }}`, context: internal.None, want: "false"},
			{name: "equaltoTrue", source: `{{ 41 is equalto(41) }}`, context: internal.None, want: "true"},
			{name: "equaltoFalse", source: `{{ 41 is equalto(42) }}`, context: internal.None, want: "false"},
			{name: "==True", source: `{{ 41 == 41 }}`, context: internal.None, want: "true"},
			{name: "==False", source: `{{ 41 == 42 }}`, context: internal.None, want: "false"},
			{name: "neTrue", source: `{{ 41 is ne(42) }}`, context: internal.None, want: "true"},
			{name: "neFalse", source: `{{ 41 is ne(41) }}`, context: internal.None, want: "false"},
			{name: "!=True", source: `{{ 41 != 42 }}`, context: internal.None, want: "true"},
			{name: "!=False", source: `{{ 41 != 41 }}`, context: internal.None, want: "false"},
			{name: "ltTrue", source: `{{ 41 is lt(42) }}`, context: internal.None, want: "true"},
			{name: "ltFalse", source: `{{ 41 is lt(41) }}`, context: internal.None, want: "false"},
			{name: "lessthanTrue", source: `{{ 41 is lessthan(42) }}`, context: internal.None, want: "true"},
			{name: "lessthanFalse", source: `{{ 41 is lessthan(41) }}`, context: internal.None, want: "false"},
			{name: "<True", source: `{{ 41 < 42 }}`, context: internal.None, want: "true"},
			{name: "<False", source: `{{ 41 < 41 }}`, context: internal.None, want: "false"},
			{name: "leTrue", source: `{{ 41 is le(41) }}`, context: internal.None, want: "true"},
			{name: "leFalse", source: `{{ 41 is le(40) }}`, context: internal.None, want: "false"},
			{name: "<=True", source: `{{ 41 <= 41 }}`, context: internal.None, want: "true"},
			{name: "<=False", source: `{{ 41 <= 40 }}`, context: internal.None, want: "false"},
			{name: "gtTrue", source: `{{ 42 is gt(41) }}`, context: internal.None, want: "true"},
			{name: "gtFalse", source: `{{ 41 is gt(41) }}`, context: internal.None, want: "false"},
			{name: "greaterthanTrue", source: `{{ 42 is greaterthan(41) }}`, context: internal.None, want: "true"},
			{name: "greaterthanFalse", source: `{{ 41 is greaterthan(41) }}`, context: internal.None, want: "false"},
			{name: ">True", source: `{{ 42 > 41 }}`, context: internal.None, want: "true"},
			{name: ">False", source: `{{ 41 > 41 }}`, context: internal.None, want: "false"},
			{name: "geTrue", source: `{{ 42 is ge(42) }}`, context: internal.None, want: "true"},
			{name: "geFalse", source: `{{ 40 is ge(41) }}`, context: internal.None, want: "false"},
			{name: ">=True", source: `{{ 41 >= 41 }}`, context: internal.None, want: "true"},
			{name: ">=False", source: `{{ 40 >= 41 }}`, context: internal.None, want: "false"},
			{name: "isInTrue", source: `{{ 1 is in([1, 2]) }}`, context: internal.None, want: "true"},
			{name: "isInFalse", source: `{{ 3 is in([1, 2]) }}`, context: internal.None, want: "false"},
			{name: "isTrueTrue", source: `{{ true is true }}`, context: internal.None, want: "true"},
			{name: "isTrueFalse", source: `{{ 1 is true }}`, context: internal.None, want: "false"},
			{name: "isFalseTrue", source: `{{ false is false }}`, context: internal.None, want: "true"},
			{name: "isFalseFalse", source: `{{ 0 is false }}`, context: internal.None, want: "false"},
			{name: "isFilterTrue", source: `{{ "escape" is filter }}`, context: internal.None, want: "true"},
			{name: "isFilterFalse", source: `{{ "no_such_filter" is filter }}`, context: internal.None, want: "false"},
			{name: "isTestTrue", source: `{{ "defined" is test }}`, context: internal.None, want: "true"},
			{name: "isTestFalse", source: `{{ "no_such_test" is test }}`, context: internal.None, want: "false"},
		})
	})
	t.Run("function", func(t *testing.T) {
		runTests(t, []testCase{
			{name: "rangeJustUpper", source: "{{ range(3) }}", context: internal.None, want: "[0, 1, 2]"},
			{name: "rangeLowerUpper", source: "{{ range(2, 4) }}", context: internal.None, want: "[2, 3]"},
			{name: "rangeLowerUpperStep", source: "{{ range(2, 9, 3) }}", context: internal.None, want: "[2, 5, 8]"},
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
			{name: "rangeNoArgErr", source: "{{ range() }}", context: internal.None, want: "missing argument"},
			{name: "rangeTooManyArgErr", source: "{{ range(1, 2, 3, 4) }}", context: internal.None, want: "too many arguments"},
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
			context: internal.None,
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
			context: internal.None,
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
			context: internal.None,
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
			context: internal.None,
			want: "\n\nHello World\n" +
				"\n\nHello 日本\n",
		}})
	})
}
