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
			context: value{typ: valueTypeMap, data: map[string]value{
				"name": {typ: valueTypeString, data: "World"},
			}},
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
			want:    "Hello <nil>", // TODO: fix
		},
		{
			source: `Hello {{ user.name }}`,
			context: value{typ: valueTypeMap, data: map[string]value{
				"user": {typ: valueTypeMap, data: map[string]value{
					"name": {typ: valueTypeString, data: "John"},
				}},
			}},
			want: "Hello John",
		},
		{
			source: `Hello {{ user["name"] }}`,
			context: value{typ: valueTypeMap, data: map[string]value{
				"user": {typ: valueTypeMap, data: map[string]value{
					"name": {typ: valueTypeString, data: "John"},
				}},
			}},
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
			context: value{typ: valueTypeMap, data: map[string]value{
				"name": {typ: valueTypeString, data: "Paul"},
			}},
			want: "Hello Paul",
		},
		{
			source:  `Hello {{ {"name": "John"}["name"] }}`,
			context: valueNone,
			want:    "Hello John",
		},
		{
			source: `Hello {{ {"name": name}["name"] }}`,
			context: value{typ: valueTypeMap, data: map[string]value{
				"name": {typ: valueTypeString, data: "Paul"},
			}},
			want: "Hello Paul",
		},
		// {
		// 	source:  `Hello {{ ["John", "Paul", "George", "Ringo"][-1] }}`,
		// 	context: valueNone,
		// 	want:    "Hello Ringo",
		// },
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
			t.Errorf("result mismatch, i=%d, source=%s, got=%s, want=%s", i, tc.source, got, tc.want)
		}
	}
}
