package mjingo

import "errors"

type Environment struct {
	syntaxConfig      SyntaxConfig
	templates         map[string]*Template
	UndefinedBehavior UndefinedBehavior
}

var ErrTemplateNotFound = errors.New("template not found")

func NewEnvironment() *Environment {
	return &Environment{
		syntaxConfig: DefaultSyntaxConfig,
		templates:    make(map[string]*Template),
	}
}

func (e *Environment) AddTemplate(name, source string) error {
	t, err := newCompiledTemplate(name, source, e.syntaxConfig)
	if err != nil {
		return err
	}
	e.templates[name] = &Template{env: e, compiled: t}
	return nil
}

func (e *Environment) GetTemplate(name string) (*Template, error) {
	tpl := e.templates[name]
	if tpl == nil {
		return nil, ErrTemplateNotFound
	}
	return tpl, nil
}
