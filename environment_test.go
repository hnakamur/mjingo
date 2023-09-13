package mjingo_test

import (
	"strings"
	"testing"

	"github.com/hnakamur/mjingo"
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
				env := mjingo.NewEnvironment()
				const templateName = "test.html"
				err := env.AddTemplate(templateName, tc.source)
				if err != nil {
					t.Fatal(err)
				}
				tpl, err := env.GetTemplate(templateName)
				if err != nil {
					t.Fatal(err)
				}
				context := mjingo.ValueFromGoValue(tc.context, mjingo.WithStructTag("json"))
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
		runTests(t, []testCase{
			{
				name:    "var",
				source:  "Hello {{ name }}",
				context: map[string]string{"name": "World"},
				want:    "Hello World",
			},
			{
				name:    "stringExpr",
				source:  `Hello {{ "world" }}`,
				context: nil,
				want:    "Hello world",
			},
			{
				name:    "i64Expr",
				source:  `Hello {{ 3 }}`,
				context: nil,
				want:    "Hello 3",
			},
			{
				name:    "f64Expr",
				source:  `Hello {{ 3.14 }}`,
				context: nil,
				want:    "Hello 3.14",
			},
			{
				name:    "boolExprTrue",
				source:  `Hello {{ true }}`,
				context: nil,
				want:    "Hello true",
			},
			{
				name:    "boolExprFalse",
				source:  `Hello {{ False }}`,
				context: nil,
				want:    "Hello false",
			},
			{
				name:    "noneExpr",
				source:  `Hello {{ none }}`,
				context: nil,
				want:    "Hello none",
			},
			{
				name:   "getFastAttr",
				source: `Hello {{ user.name }}`,
				context: map[string]any{
					"user": map[string]string{
						"name": "John",
					},
				},
				want: "Hello John",
			},
			{
				name:   "getItemOpt",
				source: `Hello {{ user["name"] }}`,
				context: map[string]any{
					"user": map[string]string{
						"name": "John",
					},
				},
				want: "Hello John",
			},
			{
				name:    "sliceString",
				source:  `Hello {{ "Johnson"[:4] }}`,
				context: nil,
				want:    "Hello John",
			},
			{
				name:    "sliceSeq",
				source:  `Hello {{ ["John", "Paul"][1] }}`,
				context: nil,
				want:    "Hello Paul",
			},
			{
				name:    "sliceVarElem",
				source:  `Hello {{ ["John", name][1] }}`,
				context: map[string]string{"name": "Paul"},
				want:    "Hello Paul",
			},
			{
				name:    "mapGetItem",
				source:  `Hello {{ {"name": "John"}["name"] }}`,
				context: nil,
				want:    "Hello John",
			},
			{
				name:    "mapVarValue",
				source:  `Hello {{ {"name": name}["name"] }}`,
				context: map[string]string{"name": "Paul"},
				want:    "Hello Paul",
			},
			{
				name:    "sliceSeqNegativeIndex",
				source:  `Hello {{ ["John", "Paul", "George", "Ringo"][-1] }}`,
				context: nil,
				want:    "Hello Ringo",
			},
			{
				name:    "addExprString",
				source:  `Hello {{ {"name": "John"}["nam" + "e"] }}`,
				context: nil,
				want:    "Hello John",
			},
			{
				name:    "addExprInt",
				source:  `Hello {{ ["John", "Paul", "George", "Ringo"][1 + 2] }}`,
				context: nil,
				want:    "Hello Ringo",
			},
			{
				name:    "addExprFloat",
				source:  `Hello {{ ["John", "Paul", "George", "Ringo"][1.0 + 2.0] }}`,
				context: nil,
				want:    "Hello Ringo",
			},
			{
				name:    "subExprInt",
				source:  `Hello {{ ["John", "Paul", "George", "Ringo"][3 - 2] }}`,
				context: nil,
				want:    "Hello Paul",
			},
			{
				name:    "subExprFloat",
				source:  `Hello {{ ["John", "Paul", "George", "Ringo"][3.0 - 2.0] }}`,
				context: nil,
				want:    "Hello Paul",
			},
			{
				name:    "stringConcat",
				source:  `{{ "Hello " ~ name ~ "!" }}`,
				context: map[string]string{"name": "John"},
				want:    "Hello John!",
			},
			{name: "pow", source: `{{ 2 ** 3 }}`, context: nil, want: "8"},
			{name: "mul", source: `{{ 2 * 3 }}`, context: nil, want: "6"},
			{name: "div", source: `{{ 3 / 2 }}`, context: nil, want: "1.5"},
			{name: "intdiv", source: `{{ 3 // 2 }}`, context: nil, want: "1"},
			{name: "rem", source: `{{ 3 % 2 }}`, context: nil, want: "1"},
			{name: "neg", source: `{{ -3 }}`, context: nil, want: "-3"},
			{name: "notTrue", source: `{{ not 0 }}`, context: nil, want: "true"},
			{name: "notFalse", source: `{{ not 1 }}`, context: nil, want: "false"},
			{name: "eq", source: `{{ 1 == 1 }}`, context: nil, want: "true"},
			{name: "lt", source: `{{ 1 < 2 }}`, context: nil, want: "true"},
			{name: "lte", source: `{{ 1 <= 1 }}`, context: nil, want: "true"},
			{name: "gt", source: `{{ 2 > 1 }}`, context: nil, want: "true"},
			{name: "gte", source: `{{ 1 >= 1 }}`, context: nil, want: "true"},
			{name: "inTrue", source: `{{ 1 in [1] }}`, context: nil, want: "true"},
			{name: "inFalse", source: `{{ 1 in [2] }}`, context: nil, want: "false"},
			{name: "inNot", source: `{{ 1 not in [2] }}`, context: nil, want: "true"},
			{name: "tuipleTreatedAsSeq0", source: `{{ () }}`, context: nil, want: "[]"},
			{name: "tuipleTreatedAsSeq1", source: `{{ (1,) }}`, context: nil, want: "[1]"},
			{name: "tuipleTreatedAsSeq2", source: `{{ (1, 2) }}`, context: nil, want: "[1, 2]"},
			{name: "scAnd1", source: `{{ false and false }}`, context: nil, want: "false"},
			{name: "scAnd2", source: `{{ true and false }}`, context: nil, want: "false"},
			{name: "scAnd3", source: `{{ true and true }}`, context: nil, want: "true"},
			{name: "scOr1", source: `{{ false or false }}`, context: nil, want: "false"},
			{name: "scOr2", source: `{{ false or true }}`, context: nil, want: "true"},
			{name: "scOr3", source: `{{ true or false }}`, context: nil, want: "true"},
		})
	})
	t.Run("statement", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:    "ifStmtNoElse",
				source:  `{% if down %}I am down{% endif %}`,
				context: map[string]any{"down": true},
				want:    "I am down",
			},
			{
				name:    "ifStmtWithElse",
				source:  `{% if down %}I am down{% else %}I am up{% endif %}`,
				context: map[string]any{"down": false},
				want:    "I am up",
			},
			{
				name:    "ifExprNoElse",
				source:  `{{ "I am down" if down }}`,
				context: map[string]any{"down": true},
				want:    "I am down",
			},
			{
				name:    "ifExprWithElse",
				source:  `{{ "I am down" if down else "I am up" }}`,
				context: map[string]any{"down": false},
				want:    "I am up",
			},
			{
				name:    "forStmtNoElse",
				source:  `{% for name in names %}{{ name }} {% endfor %}`,
				context: map[string][]string{"names": {"John", "Paul"}},
				want:    "John Paul ",
			},
			{
				name:    "forStmtWithElseUnused",
				source:  `{% for name in names %}{{ name }} {% else %}no users{% endfor %}`,
				context: map[string][]string{"names": {"John", "Paul"}},
				want:    "John Paul ",
			},
			{
				name:    "forStmtWithElseUsed",
				source:  `{% for name in names %}{{ name }} {% else %}no users{% endfor %}`,
				context: nil,
				want:    "no users",
			},
			{
				name:    "rawStmt",
				source:  `{% raw %}Hello {{ name }}{% endraw %}`,
				context: nil,
				want:    "Hello {{ name }}",
			},
			{
				name:    "withStmt",
				source:  `{% with foo = 42 %}{{ foo }}{% endwith %}`,
				context: nil,
				want:    "42",
			},
			{
				name:    "setStmt",
				source:  `{% set name = "John" %}Hello {{ name }}`,
				context: nil,
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
				context: nil,
				want: "\n" +
					"<ul>\n" +
					"\n" +
					"<li><a href=\"/\">Index</a>\n" +
					"<li><a href=\"/downloads\">Downloads</a>\n" +
					"\n" +
					"</ul>",
			},
			{
				name:    "autoEscapeStmt",
				source:  `{% autoescape "html" %}{{ unsafe }}{% endautoescape %} {% autoescape "none" %}{{ unsafe }}{% endautoescape %}`,
				context: map[string]string{"unsafe": "<foo>"},
				want:    "&lt;foo&gt; <foo>",
			},
			{
				name:    "filterStmt",
				source:  `{% filter upper %}hello{% endfilter %} world`,
				context: nil,
				want:    "HELLO world",
			},
			{
				name: "doStmt",
				source: "{% macro dialog(title) %}\n" +
					"Dialog is {{ title }}\n" +
					"{% endmacro %}\n" +
					"{% do dialog(title=\"Hello World\") %}",
				context: nil,
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
				context: nil,
				want: "\n\n  <div class=\"dialog\">\n" +
					"    <h3>Hello World</h3>\n" +
					"    <div class=\"contents\">\n  This is the dialog body.\n</div>\n" +
					"  </div>\n",
			},
		})
	})
	t.Run("loopVariable", func(t *testing.T) {
		ctx := map[string][]string{"users": {"John", "Paul", "George", "Ringo"}}

		recurCtx := map[string]any{
			"menu": []map[string]any{{
				"href":  "/menu1",
				"title": "menu1",
				"children": []map[string]any{{
					"menu": []map[string]any{{
						"href":  "/submenu1",
						"title": "submenu1",
						"children": []map[string]any{{
							"menu": []map[string]any{{
								"href":     "/submenu1-1",
								"title":    "submenu1-1",
								"children": []map[string]any{},
							}},
						}},
					}},
				}, {
					"menu": []map[string]any{{
						"href":     "/submenu2",
						"title":    "submenu2",
						"children": []map[string]any{},
					}},
				}},
			}},
		}

		changedCtx := map[string]any{
			"entries": []map[string]any{{
				"category": "Go",
				"message":  "Forward Compatibility and Toolchain Management in Go 1.21",
			}, {
				"category": "Go",
				"message":  "Structured Logging with slog",
			}, {
				"category": "Rust",
				"message":  "2022 Annual Rust Survey Results",
			}, {
				"category": "Rust",
				"message":  "Announcing Rust 1.72.0",
			}},
		}

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
			want: "<ul class=\"menu\">\n\n  <li><a href=\"&#x2f;menu1\">menu1 (depth=1)</a>\n" +
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
			want: "<ul class=\"menu\">\n\n  <li><a href=\"&#x2f;menu1\">menu1 (depth=0)</a>\n" +
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
			context: nil,
			want:    `["Hello", "World", "default", "closure"]`,
		}})
	})
	t.Run("filter", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:    "escape",
				source:  `{{ v|escape }}`,
				context: map[string]string{"v": `'foo' & "bar/baz"<br/>`},
				want:    "&#x27;foo&#x27; &amp; &quot;bar&#x2f;baz&quot;&lt;br&#x2f;&gt;",
			},
			{
				name:    "safeEscape",
				source:  `{{ v|safe|e }}`,
				context: map[string]string{"v": "<br/>"},
				want:    "<br/>",
			},
			{name: "lower", source: `{{ "HELLO"|lower }}`, context: nil, want: "hello"},
			{name: "upper", source: `{{ "hello"|upper }}`, context: nil, want: "HELLO"},
			{name: "title", source: `{{ "hello world"|title }}`, context: nil, want: "HELLO WORLD"},
			{name: "capitalize", source: `{{ "hello World"|capitalize }}`, context: nil, want: "Hello world"},
			{name: "replace", source: `{{ "Hello World"|replace("Hello", "Goodbye") }}`, context: nil, want: "Goodbye World"},
			{name: "countSlice", source: `{{ ["foo", "bar"]|length }}`, context: nil, want: "2"},
			{name: "countStr", source: `{{ "あいう"|length }}`, context: nil, want: "3"},
			{name: "dictsortCase1", source: `{% autoescape 'none' %}{{ {'a': 4, 'c': 3, 'B': 1}|dictsort }}{% endautoescape %}`, context: nil, want: `[["a", 4], ["B", 1], ["c", 3]]`},
			{name: "dictsortCase2", source: `{% autoescape 'none' %}{{ {'a': 4, 'c': 3, 'b': 1}|dictsort(by="value") }}{% endautoescape %}`, context: nil, want: `[["b", 1], ["c", 3], ["a", 4]]`},
			{name: "dictsortCase3", source: `{% autoescape 'none' %}{{ {'a': 4, 'c': 3, 'b': 1}|dictsort(by="value", reverse=true) }}{% endautoescape %}`, context: nil, want: `[["a", 4], ["c", 3], ["b", 1]]`},
			{name: "dictsortCase4", source: `{% autoescape 'none' %}{{ {'a': 4, 'c': 3, 'B': 1}|dictsort(case_sensitive=true) }}{% endautoescape %}`, context: nil, want: `[["B", 1], ["a", 4], ["c", 3]]`},
			{name: "sortCase1", source: `{% autoescape 'none' %}{{ ['a', 'c', 'B']|sort }}{% endautoescape %}`, context: nil, want: `["a", "B", "c"]`},
			{name: "sortCase2", source: `{% autoescape 'none' %}{{ ['a', 'c', 'B']|sort(reverse=true) }}{% endautoescape %}`, context: nil, want: `["c", "B", "a"]`},
			{name: "sortCase3", source: `{% autoescape 'none' %}{{ ['a', 'c', 'B']|sort(case_sensitive=true) }}{% endautoescape %}`, context: nil, want: `["B", "a", "c"]`},
			{name: "sortCase2", source: `{% autoescape 'none' %}{{ [{"name": "John", "id": 1}, {"name": "Paul", "id": 2}]|sort(attribute="id", reverse=true) }}{% endautoescape %}`, context: nil, want: `[{"name": "Paul", "id": 2}, {"name": "John", "id": 1}]`},
			{
				name:    "items",
				source:  `{% for key, value in {'a': 1, 'b': 2}|items %}{%if not loop.first %}, {% endif %}{{ key }}: {{ value }}{% endfor %}`,
				context: nil,
				want:    "a: 1, b: 2",
			},
			{name: "joinStrNoJoiner", source: `{{ "あいう"|join }}`, context: nil, want: "あいう"},
			{name: "joinStrWithJoiner", source: `{{ "あいう"|join(",") }}`, context: nil, want: "あ,い,う"},
			{name: "joinSeqNoJoiner", source: `{{ [1, 2, 3]|join }}`, context: nil, want: "123"},
			{name: "joinSeqWithJoiner", source: `{{ [1, 2, 3]|join(", ") }}`, context: nil, want: "1, 2, 3"},
			{name: "reverseStr", source: `{{ "あいう"|reverse }}`, context: nil, want: "ういあ"},
			{name: "reverseSeq", source: `{{ [1, 2, 3]|reverse }}`, context: nil, want: "[3, 2, 1]"},
			{name: "trimWithCutset", source: `{{ "¡¡¡Hello, Gophers!!!"|trim("!¡") }}`, context: nil, want: "Hello, Gophers"},
			{name: "trimNoCutset", source: `{{ " \tHello, Gophers\n "|trim }}`, context: nil, want: "Hello, Gophers"},
			{name: "defaultNoArg", source: `{{ undefined|default }}`, context: nil, want: ""},
			{name: "defaultStrArg", source: `{{ undefined|default("hello") }}`, context: nil, want: "hello"},
			{name: "defaultIntArg", source: `{{ undefined|default(2) }}`, context: nil, want: "2"},
			{name: "defaultAlias", source: `{{ undefined|d(2) }}`, context: nil, want: "2"},
			{name: "defaultAliasWithVal", source: `{{ 3|d }}`, context: nil, want: "3"},
			{name: "roundDefPrec", source: `{{ 42.5|round }}`, context: nil, want: "43.0"},
			{name: "roundWithPrec", source: `{{ 42.45|round(1) }}`, context: nil, want: "42.5"},
			{name: "absI64", source: `{{ -3|abs }}`, context: nil, want: "3"},
			{name: "absI128", source: `{{ (-9223372036854775807 * 2)|abs }}`, context: nil, want: "18446744073709551614"},
			{name: "absF64", source: `{{ -3.2|abs }}`, context: nil, want: "3.2"},
			{name: "attr", source: `{{ {'a': 1, 'b': 2}|attr('b') }}`, context: nil, want: "2"},
			{name: "firstStr", source: `{{ "あいう"|first }}`, context: nil, want: "あ"},
			{name: "firstSeq", source: `{{ [1, 2, 3]|first }}`, context: nil, want: "1"},
			{name: "lastStr", source: `{{ "あいう"|last }}`, context: nil, want: "う"},
			{name: "lastSeq", source: `{{ [1, 2, 3]|last }}`, context: nil, want: "3"},
			{name: "minSeq", source: `{{ [1, 2, 3]|min }}`, context: nil, want: "1"},
			{name: "maxSeq", source: `{{ [1, 2, 3]|max }}`, context: nil, want: "3"},
			{name: "listStr", source: `{% autoescape 'none' %}{{ "あいう"|list }}{% endautoescape %}`, context: nil, want: `["あ", "い", "う"]`},
			{name: "boolTrue", source: `{{ 1|bool }}`, context: nil, want: "true"},
			{name: "boolFalse", source: `{{ 0|bool }}`, context: nil, want: "false"},
			{name: "batchNoFiller", source: `{{ [1, 2, 3, 4, 5]|batch(3) }}`, context: nil, want: "[[1, 2, 3], [4, 5]]"},
			{name: "batchWithFiller", source: `{{ [1, 2, 3, 4, 5]|batch(3, 0) }}`, context: nil, want: "[[1, 2, 3], [4, 5, 0]]"},
			{name: "sliceNoFiller", source: `{{ [1, 2, 3, 4, 5]|slice(3) }}`, context: nil, want: "[[1, 2], [3, 4], [5]]"},
			{name: "sliceWithFiller", source: `{{ [1, 2, 3, 4, 5]|slice(3, 0) }}`, context: nil, want: "[[1, 2], [3, 4], [5, 0]]"},
			{name: "indentCase1", source: `{{ "line1\n  line2\n\n  line3\n"|indent(2) }}`, context: nil, want: "line1\n    line2\n\n    line3"},
			{name: "indentCase2", source: `{{ "line1\n  line2\n\n  line3\n"|indent(2, true) }}`, context: nil, want: "  line1\n    line2\n\n    line3"},
			{name: "indentCase3", source: `{{ "line1\n  line2\n\n  line3\n"|indent(2, false, true) }}`, context: nil, want: "line1\n    line2\n  \n    line3"},
			{name: "selectCase1", source: `{{ [1, 2, 3, 4, 5]|select("odd") }}`, context: nil, want: "[1, 3, 5]"},
			{name: "rejectCase1", source: `{{ [1, 2, 3, 4, 5]|reject("odd") }}`, context: nil, want: "[2, 4]"},
			{name: "selectattrCase1", source: `{% autoescape 'none' %}{{ [{"name": "John", "is_active": false}, {"name": "Paul", "is_active": true}]|selectattr("is_active") }}{% endautoescape %}`, context: nil, want: `[{"name": "Paul", "is_active": true}]`},
			{name: "selectattrCase2", source: `{% autoescape 'none' %}{{ [{"name": "John", "id": 1}, {"name": "Paul", "id": 2}]|selectattr("id", "even") }}{% endautoescape %}`, context: nil, want: `[{"name": "Paul", "id": 2}]`},
			{name: "rejectattrCase1", source: `{% autoescape 'none' %}{{ [{"name": "John", "is_active": false}, {"name": "Paul", "is_active": true}]|rejectattr("is_active") }}{% endautoescape %}`, context: nil, want: `[{"name": "John", "is_active": false}]`},
			{name: "rejectattrCase2", source: `{% autoescape 'none' %}{{ [{"name": "John", "id": 1}, {"name": "Paul", "id": 2}]|rejectattr("id", "even") }}{% endautoescape %}`, context: nil, want: `[{"name": "John", "id": 1}]`},
			{name: "unique", source: `{% autoescape 'none' %}{{ ['foo', 'bar', 'foobar', 'foobar']|unique }}{% endautoescape %}`, context: nil, want: `["foo", "bar", "foobar"]`},
			{name: "mapCase1", source: `{% autoescape 'none' %}{{ [{"name": "John", "id": 1}, {"name": "Paul", "id": 2}]|map(attribute="name")|join(', ') }}{% endautoescape %}`, context: nil, want: `John, Paul`},
			{name: "mapCase2", source: `{% autoescape 'none' %}{{ [-1, -2, 3, 4, -5]|map("abs")|join(", ") }}{% endautoescape %}`, context: nil, want: `1, 2, 3, 4, 5`},
		})
	})
	t.Run("test", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:    "isDefined",
				source:  `{% if v is defined %}I am defined{% else %}I am fallback{% endif %}`,
				context: map[string]any{"v": nil},
				want:    "I am defined",
			},
			{
				name:    "isNotDefined",
				source:  `{% if v is not defined %}I am fallback{% endif %}`,
				context: nil,
				want:    "I am fallback",
			},
			{
				name:    "isNone",
				source:  `{% if v is none %}I am none{% else %}I am not none{% endif %}`,
				context: map[string]any{"v": nil},
				want:    "I am none",
			},
			{
				name:    "isSafeTrue",
				source:  `{% if v is safe %}I am safe{% else %}I am not safe{% endif %}`,
				context: map[string]any{"v": mjingo.ValueFromSafeString("s")},
				want:    "I am safe",
			},
			{
				name:    "isSafeFalse",
				source:  `{% if v is safe %}I am safe{% else %}I am not safe{% endif %}`,
				context: map[string]any{"v": "s"},
				want:    "I am not safe",
			},
			{
				name:    "isEscaped",
				source:  `{% if v is escaped %}I am safe{% else %}I am not safe{% endif %}`,
				context: map[string]any{"v": mjingo.ValueFromSafeString("s")},
				want:    "I am safe",
			},
			{name: "isOddTrue", source: `{{ 41 is odd }}`, context: nil, want: "true"},
			{name: "isOddValueFalse", source: `{{ 42 is odd }}`, context: nil, want: "false"},
			{name: "isOddTypeFalse", source: `{{ "s" is odd }}`, context: nil, want: "false"},
			{name: "isEvenTrue", source: `{{ 41 is even }}`, context: nil, want: "false"},
			{name: "isEvenValueFalse", source: `{{ 42 is even }}`, context: nil, want: "true"},
			{name: "isEvenTypeFalse", source: `{{ "s" is even }}`, context: nil, want: "false"},
			{name: "isNumberTrue", source: `{{ 42 is number }}`, context: nil, want: "true"},
			{name: "isNumberFalse", source: `{{ "42" is number }}`, context: nil, want: "false"},
			{name: "isStringTrue", source: `{{ "42" is string }}`, context: nil, want: "true"},
			{name: "isStringFalse", source: `{{ 42 is string }}`, context: nil, want: "false"},
			{name: "isSequenceTrue", source: `{{ [1, 2, 3] is sequence }}`, context: nil, want: "true"},
			{name: "isSequenceFalse", source: `{{ 42 is sequence }}`, context: nil, want: "false"},
			{name: "isMappingTrue", source: `{{ {"foo": "bar"} is mapping }}`, context: nil, want: "true"},
			{name: "isMappingFalse", source: `{{ [1, 2, 3] is mapping }}`, context: nil, want: "false"},
			{name: "isStartingWithTrue", source: `{{ "foobar" is startingwith("foo") }}`, context: nil, want: "true"},
			{name: "isStartingWithFalse", source: `{{ "foobar" is startingwith("bar") }}`, context: nil, want: "false"},
			{name: "isEndingWithTrue", source: `{{ "foobar" is endingwith("bar") }}`, context: nil, want: "true"},
			{name: "isEndingWithFalse", source: `{{ "foobar" is endingwith("foo") }}`, context: nil, want: "false"},
			{name: "eqTrue", source: `{{ 41 is eq(41) }}`, context: nil, want: "true"},
			{name: "eqFalse", source: `{{ 41 is eq(42) }}`, context: nil, want: "false"},
			{name: "equaltoTrue", source: `{{ 41 is equalto(41) }}`, context: nil, want: "true"},
			{name: "equaltoFalse", source: `{{ 41 is equalto(42) }}`, context: nil, want: "false"},
			{name: "==True", source: `{{ 41 == 41 }}`, context: nil, want: "true"},
			{name: "==False", source: `{{ 41 == 42 }}`, context: nil, want: "false"},
			{name: "neTrue", source: `{{ 41 is ne(42) }}`, context: nil, want: "true"},
			{name: "neFalse", source: `{{ 41 is ne(41) }}`, context: nil, want: "false"},
			{name: "!=True", source: `{{ 41 != 42 }}`, context: nil, want: "true"},
			{name: "!=False", source: `{{ 41 != 41 }}`, context: nil, want: "false"},
			{name: "ltTrue", source: `{{ 41 is lt(42) }}`, context: nil, want: "true"},
			{name: "ltFalse", source: `{{ 41 is lt(41) }}`, context: nil, want: "false"},
			{name: "lessthanTrue", source: `{{ 41 is lessthan(42) }}`, context: nil, want: "true"},
			{name: "lessthanFalse", source: `{{ 41 is lessthan(41) }}`, context: nil, want: "false"},
			{name: "<True", source: `{{ 41 < 42 }}`, context: nil, want: "true"},
			{name: "<False", source: `{{ 41 < 41 }}`, context: nil, want: "false"},
			{name: "leTrue", source: `{{ 41 is le(41) }}`, context: nil, want: "true"},
			{name: "leFalse", source: `{{ 41 is le(40) }}`, context: nil, want: "false"},
			{name: "<=True", source: `{{ 41 <= 41 }}`, context: nil, want: "true"},
			{name: "<=False", source: `{{ 41 <= 40 }}`, context: nil, want: "false"},
			{name: "gtTrue", source: `{{ 42 is gt(41) }}`, context: nil, want: "true"},
			{name: "gtFalse", source: `{{ 41 is gt(41) }}`, context: nil, want: "false"},
			{name: "greaterthanTrue", source: `{{ 42 is greaterthan(41) }}`, context: nil, want: "true"},
			{name: "greaterthanFalse", source: `{{ 41 is greaterthan(41) }}`, context: nil, want: "false"},
			{name: ">True", source: `{{ 42 > 41 }}`, context: nil, want: "true"},
			{name: ">False", source: `{{ 41 > 41 }}`, context: nil, want: "false"},
			{name: "geTrue", source: `{{ 42 is ge(42) }}`, context: nil, want: "true"},
			{name: "geFalse", source: `{{ 40 is ge(41) }}`, context: nil, want: "false"},
			{name: ">=True", source: `{{ 41 >= 41 }}`, context: nil, want: "true"},
			{name: ">=False", source: `{{ 40 >= 41 }}`, context: nil, want: "false"},
			{name: "isInTrue", source: `{{ 1 is in([1, 2]) }}`, context: nil, want: "true"},
			{name: "isInFalse", source: `{{ 3 is in([1, 2]) }}`, context: nil, want: "false"},
			{name: "isTrueTrue", source: `{{ true is true }}`, context: nil, want: "true"},
			{name: "isTrueFalse", source: `{{ 1 is true }}`, context: nil, want: "false"},
			{name: "isFalseTrue", source: `{{ false is false }}`, context: nil, want: "true"},
			{name: "isFalseFalse", source: `{{ 0 is false }}`, context: nil, want: "false"},
			{name: "isFilterTrue", source: `{{ "escape" is filter }}`, context: nil, want: "true"},
			{name: "isFilterFalse", source: `{{ "no_such_filter" is filter }}`, context: nil, want: "false"},
			{name: "isTestTrue", source: `{{ "defined" is test }}`, context: nil, want: "true"},
			{name: "isTestFalse", source: `{{ "no_such_test" is test }}`, context: nil, want: "false"},
		})
	})
	t.Run("function", func(t *testing.T) {
		runTests(t, []testCase{
			{name: "rangeJustUpper", source: "{{ range(3) }}", context: nil, want: "[0, 1, 2]"},
			{name: "rangeLowerUpper", source: "{{ range(2, 4) }}", context: nil, want: "[2, 3]"},
			{name: "rangeLowerUpperStep", source: "{{ range(2, 9, 3) }}", context: nil, want: "[2, 5, 8]"},
			{name: "dict", source: "{{ dict(foo='bar')['foo'] }}", context: nil, want: "bar"},
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
				env := mjingo.NewEnvironment()
				const templateName = "test.html"
				err := env.AddTemplate(templateName, tc.source)
				if err != nil {
					t.Fatal(err)
				}
				tpl, err := env.GetTemplate(templateName)
				if err != nil {
					t.Fatal(err)
				}
				context := mjingo.ValueFromGoValue(tc.context, mjingo.WithStructTag("json"))
				var got string
				if _, err := tpl.Render(context); err != nil {
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
			{name: "rangeNoArgErr", source: "{{ range() }}", context: nil, want: "missing argument"},
			{name: "rangeTooManyArgErr", source: "{{ range(1, 2, 3, 4) }}", context: nil, want: "too many arguments"},
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
				env := mjingo.NewEnvironment()
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
				context := mjingo.ValueFromGoValue(tc.context, mjingo.WithStructTag("json"))
				got, err := tpl.Render(context)
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
			context: nil,
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
			context: nil,
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
			context: nil,
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
			context: nil,
			want: "\n\nHello World\n" +
				"\n\nHello 日本\n",
		}})
	})
}
func TestExpression(t *testing.T) {
	type testCase struct {
		name    string
		expr    string
		context any
		want    mjingo.Value
	}

	runTests := func(t *testing.T, testCases []testCase) {
		t.Helper()
		for i, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				env := mjingo.NewEnvironment()
				expr, err := env.CompileExpression(tc.expr)
				if err != nil {
					t.Fatal(err)
				}
				root := mjingo.ValueFromGoValue(tc.context, mjingo.WithStructTag("json"))
				got, err := expr.Eval(root)
				if err != nil {
					t.Fatal(err)
				}
				if !got.Equal(tc.want) {
					t.Errorf("result mismatch, i=%d, expr=%s,\n got=%q,\nwant=%q", i, tc.expr, got, tc.want)
				}
			})
		}
	}

	runTests(t, []testCase{{
		name: "add",
		expr: "foo + bar",
		context: map[string]int{
			"foo": 42,
			"bar": 23,
		},
		want: mjingo.ValueFromGoValue(65),
	}})
}

func TestKeepTrailingNewline(t *testing.T) {
	env := mjingo.NewEnvironment()
	env.SetKeepTrailingNewline(true)
	const templateName = "test.html"
	const source = "Hello {{ name }}\n"
	err := env.AddTemplate(templateName, source)
	if err != nil {
		t.Fatal(err)
	}
	tpl, err := env.GetTemplate(templateName)
	if err != nil {
		t.Fatal(err)
	}
	context := mjingo.ValueFromGoValue(map[string]any{"name": "World"}, mjingo.WithStructTag("json"))
	got, err := tpl.Render(context)
	if err != nil {
		t.Fatal(err)
	}
	const want = "Hello World\n"
	if got != want {
		t.Errorf("result mismatch, source=%s,\n got=%q,\nwant=%q", source, got, want)
	}
}

func TestRenderNamedStr(t *testing.T) {
	env := mjingo.NewEnvironment()
	context := mjingo.ValueFromGoValue(map[string]any{"name": "World"}, mjingo.WithStructTag("json"))
	got, err := env.RenderNamedStr("hello", "Hello {{ name }}\n", context)
	if err != nil {
		t.Fatal(err)
	}
	const want = "Hello World"
	if got != want {
		t.Errorf("result mismatch, got=%q, want=%q", got, want)
	}
}

func TestRenderStr(t *testing.T) {
	env := mjingo.NewEnvironment()
	context := mjingo.ValueFromGoValue(map[string]any{"name": "World"}, mjingo.WithStructTag("json"))
	got, err := env.RenderStr("Hello {{ name }}\n", context)
	if err != nil {
		t.Fatal(err)
	}
	const want = "Hello World"
	if got != want {
		t.Errorf("result mismatch, got=%q, want=%q", got, want)
	}
}

func TestSetAutoEscapeCallback(t *testing.T) {
	env := mjingo.NewEnvironment()
	env.SetAutoEscapeCallback(func(_name string) mjingo.AutoEscape {
		return mjingo.AutoEscapeNone
	})
	const templateName = "test.html"
	const source = "Hello {{ name }}"
	err := env.AddTemplate(templateName, source)
	if err != nil {
		t.Fatal(err)
	}
	tpl, err := env.GetTemplate(templateName)
	if err != nil {
		t.Fatal(err)
	}
	context := mjingo.ValueFromGoValue(map[string]any{"name": "A & B"}, mjingo.WithStructTag("json"))
	got, err := tpl.Render(context)
	if err != nil {
		t.Fatal(err)
	}
	const want = "Hello A & B"
	if got != want {
		t.Errorf("result mismatch, source=%s,\n got=%q,\nwant=%q", source, got, want)
	}
}

func TestSetUndefinedBehavior(t *testing.T) {
	env := mjingo.NewEnvironment()
	env.SetUndefinedBehavior(mjingo.UndefinedBehaviorChainable)
	const templateName = "test.html"
	const source = "Hello {{ name }}"
	err := env.AddTemplate(templateName, source)
	if err != nil {
		t.Fatal(err)
	}
	tpl, err := env.GetTemplate(templateName)
	if err != nil {
		t.Fatal(err)
	}
	context := mjingo.ValueFromGoValue(map[string]any{})
	got, err := tpl.Render(context)
	if err != nil {
		t.Fatal(err)
	}
	const want = "Hello "
	if got != want {
		t.Errorf("result mismatch, source=%s,\n got=%q,\nwant=%q", source, got, want)
	}
}

func TestEnvironment_AddFilter_withStateArg(t *testing.T) {
	slugify := func(_ *mjingo.State, value string) string {
		return strings.Join(strings.Split(strings.ToLower(value), " "), "-")
	}

	env := mjingo.NewEnvironment()
	env.AddFilter("slugify", mjingo.BoxedFilterFromFuncReflect(slugify))
	const templateName = "test.txt"
	err := env.AddTemplate(templateName, `{{ title|slugify }}`)
	if err != nil {
		t.Fatal(err)
	}
	tpl, err := env.GetTemplate(templateName)
	if err != nil {
		t.Fatal(err)
	}
	context := mjingo.ValueFromGoValue(map[string]string{"title": "this is my page"})
	got, err := tpl.Render(context)
	if err != nil {
		t.Fatal(err)
	}
	const want = "this-is-my-page"
	if got != want {
		t.Errorf("result mismatch, got=%q,\nwant=%q", got, want)
	}
}
