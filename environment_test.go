package mjingo

import (
	"testing"
)

func TestEnvironment(t *testing.T) {
	testCases := []struct {
		source  string
		context any
		want    string
	}{
		{
			source: "Hello {{ name }}",
			context: valueFromValueIndexMap(valueIndexMapFromKeyRefValues([]keyRefValue{{
				key: keyRefFromString("name"), val: valueFromString("World"),
			}})),
			want: "Hello World",
		},
		{
			source:  `Hello {{ "world" }}`,
			context: valueNone,
			want:    "Hello world",
		},
		{
			source:  `Hello {{ 3 }}`,
			context: valueNone,
			want:    "Hello 3",
		},
		{
			source:  `Hello {{ 3.14 }}`,
			context: valueNone,
			want:    "Hello 3.14",
		},
		{
			source:  `Hello {{ true }}`,
			context: valueNone,
			want:    "Hello true",
		},
		{
			source:  `Hello {{ False }}`,
			context: valueNone,
			want:    "Hello false",
		},
		{
			source:  `Hello {{ none }}`,
			context: valueNone,
			want:    "Hello none",
		},
		{
			source: `Hello {{ user.name }}`,
			context: valueFromValueIndexMap(valueIndexMapFromKeyRefValues([]keyRefValue{{
				key: keyRefFromString("user"),
				val: valueFromValueIndexMap(valueIndexMapFromKeyRefValues([]keyRefValue{{
					key: keyRefFromString("name"),
					val: valueFromString("John"),
				}})),
			}})),
			want: "Hello John",
		},
		{
			source: `Hello {{ user["name"] }}`,
			context: valueFromValueIndexMap(valueIndexMapFromKeyRefValues([]keyRefValue{{
				key: keyRefFromString("user"),
				val: valueFromValueIndexMap(valueIndexMapFromKeyRefValues([]keyRefValue{{
					key: keyRefFromValue(valueFromString("name")),
					val: valueFromString("John"),
				}})),
			}})),
			want: "Hello John",
		},
		{
			source:  `Hello {{ "Johnson"[:4] }}`,
			context: valueNone,
			want:    "Hello John",
		},
		{
			source:  `Hello {{ ["John", "Paul"][1] }}`,
			context: valueNone,
			want:    "Hello Paul",
		},
		{
			source: `Hello {{ ["John", name][1] }}`,
			context: valueFromValueIndexMap(valueIndexMapFromKeyRefValues([]keyRefValue{{
				key: keyRefFromString("name"), val: valueFromString("Paul"),
			}})),
			want: "Hello Paul",
		},
		{
			source:  `Hello {{ {"name": "John"}["name"] }}`,
			context: valueNone,
			want:    "Hello John",
		},
		{
			source: `Hello {{ {"name": name}["name"] }}`,
			context: valueFromValueIndexMap(valueIndexMapFromKeyRefValues([]keyRefValue{{
				key: keyRefFromString("name"), val: valueFromString("Paul"),
			}})),
			want: "Hello Paul",
		},
		{
			source:  `Hello {{ ["John", "Paul", "George", "Ringo"][-1] }}`,
			context: valueNone,
			want:    "Hello Ringo",
		},
		{
			source:  `Hello {{ {"name": "John"}["nam" + "e"] }}`,
			context: valueNone,
			want:    "Hello John",
		},
		{
			source:  `Hello {{ ["John", "Paul", "George", "Ringo"][1 + 2] }}`,
			context: valueNone,
			want:    "Hello Ringo",
		},
		{
			source:  `Hello {{ ["John", "Paul", "George", "Ringo"][1.0 + 2.0] }}`,
			context: valueNone,
			want:    "Hello Ringo",
		},
		{
			source:  `Hello {{ ["John", "Paul", "George", "Ringo"][3 - 2] }}`,
			context: valueNone,
			want:    "Hello Paul",
		},
		{
			source:  `Hello {{ ["John", "Paul", "George", "Ringo"][3.0 - 2.0] }}`,
			context: valueNone,
			want:    "Hello Paul",
		},
		{
			source: `{{ "Hello " ~ name ~ "!" }}`,
			context: valueFromValueIndexMap(valueIndexMapFromKeyRefValues([]keyRefValue{{
				key: keyRefFromString("name"), val: valueFromString("John"),
			}})),
			want: "Hello John!",
		},
		{
			source:  `{{ 2 ** 3 }}`,
			context: valueNone,
			want:    "8",
		},
		{
			source: `{% if down %}I'm down{% endif %}`,
			context: valueFromValueIndexMap(valueIndexMapFromKeyRefValues([]keyRefValue{{
				key: keyRefFromString("down"), val: valueFromBool(true),
			}})),
			want: "I'm down",
		},
		{
			source: `{% if down %}I'm down{% else %}I'm up{% endif %}`,
			context: valueFromValueIndexMap(valueIndexMapFromKeyRefValues([]keyRefValue{{
				key: keyRefFromString("down"), val: valueFromBool(false),
			}})),
			want: "I'm up",
		},
		{
			source: `{{ "I'm down" if down }}`,
			context: valueFromValueIndexMap(valueIndexMapFromKeyRefValues([]keyRefValue{{
				key: keyRefFromString("down"), val: valueFromBool(true),
			}})),
			want: "I'm down",
		},
		{
			source: `{{ "I'm down" if down else "I'm up" }}`,
			context: valueFromValueIndexMap(valueIndexMapFromKeyRefValues([]keyRefValue{{
				key: keyRefFromString("down"), val: valueFromBool(false),
			}})),
			want: "I'm up",
		},
		{
			source: `{% for name in names %}{{ name }} {% endfor %}`,
			context: valueFromValueIndexMap(valueIndexMapFromKeyRefValues([]keyRefValue{{
				key: keyRefFromString("names"), val: valueFromValueSlice([]value{
					valueFromString("John"),
					valueFromString("Paul"),
				}),
			}})),
			want: "John Paul ",
		},
		{
			source:  `{% raw %}Hello {{ name }}{% endraw %}`,
			context: valueNone,
			want:    "Hello {{ name }}",
		},
		{
			source:  `{% with foo = 42 %}{{ foo }}{% endwith %}`,
			context: valueNone,
			want:    "42",
		},
		{
			source:  `{% set name = "John" %}Hello {{ name }}`,
			context: valueNone,
			want:    "Hello John",
		},
		{
			source: "{% set navigation %}\n" +
				"<li><a href=\"/\">Index</a>\n" +
				"<li><a href=\"/downloads\">Downloads</a>\n" +
				"{% endset %}\n" +
				"<ul>\n" +
				"{{ navigation }}\n" +
				"</ul>\n",
			context: valueNone,
			want: "\n" +
				"<ul>\n" +
				"\n" +
				"<li><a href=\"/\">Index</a>\n" +
				"<li><a href=\"/downloads\">Downloads</a>\n" +
				"\n" +
				"</ul>",
		},
		{
			source:  `{% if seq is not defined %}I'm fallback{% endif %}`,
			context: valueNone,
			want:    "I'm fallback",
		},
	}
	for i, tc := range testCases {
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
	}
}
