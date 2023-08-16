package mjingo

import (
	"errors"
	"io"
	"strings"
)

type Environment struct {
	syntaxConfig      SyntaxConfig
	templates         map[string]*Template
	tests             map[string]TestFunc
	globals           map[string]value
	defaultAutoEscape autoEscapeCallBack
	UndefinedBehavior UndefinedBehavior
}

type autoEscapeCallBack func(name string) autoEscape

var ErrTemplateNotFound = errors.New("template not found")

func NewEnvironment() *Environment {
	return &Environment{
		syntaxConfig:      DefaultSyntaxConfig,
		templates:         make(map[string]*Template),
		tests:             make(map[string]TestFunc),
		globals:           make(map[string]value),
		defaultAutoEscape: defaultAutoEscapeCallback,
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
	return &Template{
		env:               e,
		compiled:          tpl.compiled,
		initialAutoEscape: e.initialAutoEscape(name),
	}, nil
}

func (e *Environment) format(v value, state *virtualMachineState, out io.Writer) error {
	if v.isUndefined() && e.UndefinedBehavior == UndefinedBehaviorStrict {
		return &Error{typ: UndefinedError}
	}
	// TODO: use formatter
	if _, err := io.WriteString(out, v.String()); err != nil {
		return err
	}
	return nil
}

func (e *Environment) getGlobal(name string) option[value] {
	val := e.globals[name]
	if val != nil {
		return option[value]{valid: true, data: val.clone()}
	}
	return option[value]{}
}

func (e *Environment) initialAutoEscape(name string) autoEscape {
	return e.defaultAutoEscape(name)
}

func (e *Environment) getTest(name string) option[TestFunc] {
	if f, ok := e.tests[name]; ok {
		return option[TestFunc]{valid: true, data: f}
	}
	return option[TestFunc]{}
}

func noAutoEscape(_ string) autoEscape { return autoEscapeNone{} }

func defaultAutoEscapeCallback(name string) autoEscape {
	_, suffix, found := strings.Cut(name, ".")
	if found {
		switch suffix {
		case "html", "htm", "xml":
			return autoEscapeHTML{}
		case "json", "json5", "js", "yaml", "yml":
			return autoEscapeJSON{}
		}
	}
	return autoEscapeNone{}
}
