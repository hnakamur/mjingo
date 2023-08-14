package mjingo

import (
	"log"
	"testing"
)

func TestEnvironment(t *testing.T) {
	testCases := []struct {
		source  string
		context any
	}{
		{
			source: "Hello {{ name }}",
			context: value{kind: valueKindMap, data: map[string]value{
				"name": {kind: valueKindString, data: "World"},
			}},
		},
		{
			source: `Hello {{ "world" }}`,
			context: value{kind: valueKindMap, data: map[string]value{
				"name": {kind: valueKindString, data: "World"},
			}},
		},
		{
			source: `Hello {{ 3 }}`,
			context: value{kind: valueKindMap, data: map[string]value{
				"name": {kind: valueKindString, data: "World"},
			}},
		},
		{
			source: `Hello {{ 3.14 }}`,
			context: value{kind: valueKindMap, data: map[string]value{
				"name": {kind: valueKindString, data: "World"},
			}},
		},
		{
			source: `Hello {{ true }}`,
			context: value{kind: valueKindMap, data: map[string]value{
				"name": {kind: valueKindString, data: "World"},
			}},
		},
		{
			source: `Hello {{ False }}`,
			context: value{kind: valueKindMap, data: map[string]value{
				"name": {kind: valueKindString, data: "World"},
			}},
		},
		{
			source: `Hello {{ none }}`,
			context: value{kind: valueKindMap, data: map[string]value{
				"name": {kind: valueKindString, data: "World"},
			}},
		},
		{
			source: `Hello {{ user.name }}`,
			context: value{kind: valueKindMap, data: map[string]value{
				"user": {kind: valueKindMap, data: map[string]value{
					"name": {kind: valueKindString, data: "John"},
				}},
			}},
		},
		{
			source: `Hello {{ user["name"] }}`,
			context: value{kind: valueKindMap, data: map[string]value{
				"user": {kind: valueKindMap, data: map[string]value{
					"name": {kind: valueKindString, data: "John"},
				}},
			}},
		},
		{
			source:  `Hello {{ "Johnson"[:4] }}`,
			context: value{kind: valueKindNone},
		},
	}
	for _, tc := range testCases {
		env := NewEnvironment()
		const templateName = "foo.js"
		log.Printf("calling AddTemplate source=%s", tc.source)
		err := env.AddTemplate(templateName, tc.source)
		if err != nil {
			t.Fatal(err)
		}
		tpl, err := env.GetTemplate(templateName)
		if err != nil {
			t.Fatal(err)
		}
		output, err := tpl.render(tc.context)
		if err != nil {
			t.Fatal(err)
		}
		log.Printf("source=%s, output=%q", tc.source, output)
	}
}
