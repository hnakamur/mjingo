package mjingo

import (
	"strings"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

type Environment struct {
	syntaxConfig      SyntaxConfig
	templates         map[string]*Template
	filters           map[string]BoxedFilter
	tests             map[string]BoxedTest
	globals           map[string]Value
	defaultAutoEscape autoEscapeFunc
	undefinedBehavior UndefinedBehavior
	formatter         formatterFunc
}

type autoEscapeFunc func(name string) AutoEscape
type formatterFunc = func(*output, *vmState, Value) error

func NewEnvironment() *Environment {
	return &Environment{
		syntaxConfig:      DefaultSyntaxConfig,
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
		return nil, newError(TemplateNotFound, "")
	}
	return &Template{
		env:               e,
		compiled:          tpl.compiled,
		initialAutoEscape: e.initialAutoEscape(name),
	}, nil
}

func (e *Environment) format(v Value, state *vmState, out *output) error {
	if v.isUndefined() && e.undefinedBehavior == UndefinedBehaviorStrict {
		return newError(UndefinedError, "")
	}
	return e.formatter(out, state, v)
}

func (e *Environment) getGlobal(name string) option.Option[Value] {
	val := e.globals[name]
	if val != nil {
		return option.Some(val.clone())
	}
	return option.None[Value]()
}

func (e *Environment) initialAutoEscape(name string) AutoEscape {
	return e.defaultAutoEscape(name)
}

func (e *Environment) getFilter(name string) option.Option[BoxedFilter] {
	if f, ok := e.filters[name]; ok {
		return option.Some(f)
	}
	return option.None[BoxedFilter]()
}

func (e *Environment) getTest(name string) option.Option[BoxedTest] {
	if f, ok := e.tests[name]; ok {
		return option.Some(f)
	}
	return option.None[BoxedTest]()
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
