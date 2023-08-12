package mjingo

import "errors"

type Environment struct {
	templates map[string]*Template
}

type Template struct {
}

var ErrTemplateNotFound = errors.New("template not found")

func (e *Environment) AddTemplate(name, template string) error {
	return nil
}

func (e *Environment) GetTemplate(name string) (*Template, error) {
	tpl := e.templates[name]
	if tpl == nil {
		return nil, ErrTemplateNotFound
	}
	return tpl, nil
}

func (t *Template) Render(context any) (string, error) {
	return "", nil
}
