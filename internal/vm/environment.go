package vm

import (
	"errors"
	"strings"

	"github.com/hnakamur/mjingo/internal/common"
	"github.com/hnakamur/mjingo/internal/compiler"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/internal/value"
)

type Environment struct {
	syntaxConfig      compiler.SyntaxConfig
	templates         map[string]*Template
	filters           map[string]FilterFunc
	tests             map[string]TestFunc
	globals           map[string]value.Value
	defaultAutoEscape autoEscapeFunc
	undefinedBehavior UndefinedBehavior
	formatter         formatterFunc
}

type autoEscapeFunc func(name string) AutoEscape
type formatterFunc = func(*Output, *State, value.Value) error

var ErrTemplateNotFound = errors.New("template not found")

func NewEnvironment() *Environment {
	return &Environment{
		syntaxConfig:      compiler.DefaultSyntaxConfig,
		templates:         make(map[string]*Template),
		filters:           getDefaultBuiltinFilters(),
		tests:             getDefaultBuiltinTests(),
		globals:           getDefaultGlobals(),
		defaultAutoEscape: defaultAutoEscapeCallback,
		formatter:         escapeFormatter,
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

func (e *Environment) format(v value.Value, state *State, out *Output) error {
	if v.IsUndefined() && e.undefinedBehavior == UndefinedBehaviorStrict {
		return common.NewError(common.UndefinedError, "")
	}
	return e.formatter(out, state, v)
}

func (e *Environment) getGlobal(name string) option.Option[value.Value] {
	val := e.globals[name]
	if val != nil {
		return option.Some(val.Clone())
	}
	return option.None[value.Value]()
}

func (e *Environment) initialAutoEscape(name string) AutoEscape {
	return e.defaultAutoEscape(name)
}

func (e *Environment) getFilter(name string) option.Option[FilterFunc] {
	if f, ok := e.filters[name]; ok {
		return option.Some(f)
	}
	return option.None[FilterFunc]()
}

func (e *Environment) getTest(name string) option.Option[TestFunc] {
	if f, ok := e.tests[name]; ok {
		return option.Some(f)
	}
	return option.None[TestFunc]()
}

func noAutoEscape(_ string) AutoEscape { return AutoEscapeNone{} }

func defaultAutoEscapeCallback(name string) AutoEscape {
	_, suffix, found := strings.Cut(name, ".")
	if found {
		switch suffix {
		case "html", "htm", "xml":
			return AutoEscapeHTML{}
		case "json", "json5", "js", "yaml", "yml":
			return AutoEscapeJSON{}
		}
	}
	return AutoEscapeNone{}
}
