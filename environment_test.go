package mjingo

import (
	"log"
	"testing"
)

func TestEnvironment(t *testing.T) {
	env := NewEnvironment()
	const templateName = "foo.js"
	err := env.AddTemplate(templateName, "Hello {{ name }}")
	if err != nil {
		t.Fatal(err)
	}
	tpl, err := env.GetTemplate(templateName)
	if err != nil {
		t.Fatal(err)
	}
	ctx := value{kind: valueKindMap, data: map[string]value{
		"name": {kind: valueKindString, data: "World"},
	}}
	output, err := tpl.render(ctx)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("output=%q", output)
}
