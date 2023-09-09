package mjingo_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/hnakamur/mjingo"
)

func ExampleEnvironment_AddFilter() {
	slugify := func(value string) string {
		return strings.Join(strings.Split(strings.ToLower(value), " "), "-")
	}

	env := mjingo.NewEnvironment()
	env.AddFilter("slugify", slugify)
	const templateName = "test.txt"
	err := env.AddTemplate(templateName, `{{ title|slugify }}`)
	if err != nil {
		log.Fatal(err)
	}
	tpl, err := env.GetTemplate(templateName)
	if err != nil {
		log.Fatal(err)
	}
	context := mjingo.ValueFromGoValue(map[string]string{"title": "this is my page"})
	got, err := tpl.Render(context)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(got)
	// Output: this-is-my-page
}
