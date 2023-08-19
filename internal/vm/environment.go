package vm

import (
	"errors"
	"io"
	"strings"

	"github.com/hnakamur/mjingo/internal"
	"github.com/hnakamur/mjingo/internal/compiler"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/value"
)

type Environment struct {
	syntaxConfig      compiler.SyntaxConfig
	templates         map[string]*Template
	tests             map[string]TestFunc
	globals           map[string]value.Value
	defaultAutoEscape autoEscapeCallBack
	undefinedBehavior compiler.UndefinedBehavior
}

type autoEscapeCallBack func(name string) compiler.AutoEscape

var ErrTemplateNotFound = errors.New("template not found")

func NewEnvironment() *Environment {
	return &Environment{
		syntaxConfig:      compiler.DefaultSyntaxConfig,
		templates:         make(map[string]*Template),
		tests:             getDefaultBuiltinTests(),
		globals:           make(map[string]value.Value),
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

func (e *Environment) format(v value.Value, state *State, out io.Writer) error {
	if v.IsUndefined() && e.undefinedBehavior == compiler.UndefinedBehaviorStrict {
		return internal.NewError(internal.UndefinedError, "")
	}
	// TODO: use formatter
	if _, err := io.WriteString(out, v.String()); err != nil {
		return err
	}
	return nil
}

func (e *Environment) getGlobal(name string) option.Option[value.Value] {
	val := e.globals[name]
	if val != nil {
		return option.Some(val.Clone())
	}
	return option.None[value.Value]()
}

func (e *Environment) initialAutoEscape(name string) compiler.AutoEscape {
	return e.defaultAutoEscape(name)
}

func (e *Environment) getTest(name string) option.Option[TestFunc] {
	if f, ok := e.tests[name]; ok {
		return option.Some(f)
	}
	return option.None[TestFunc]()
}

func noAutoEscape(_ string) compiler.AutoEscape { return compiler.AutoEscapeNone{} }

func defaultAutoEscapeCallback(name string) compiler.AutoEscape {
	_, suffix, found := strings.Cut(name, ".")
	if found {
		switch suffix {
		case "html", "htm", "xml":
			return compiler.AutoEscapeHTML{}
		case "json", "json5", "js", "yaml", "yml":
			return compiler.AutoEscapeJSON{}
		}
	}
	return compiler.AutoEscapeNone{}
}
