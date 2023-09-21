package mjingo

import (
	"github.com/hnakamur/mjingo/option"
)

// Environment is an abstraction that holds the engine configuration.
//
// This object holds the central configuration state for templates.  It is also
// the container for all loaded templates.
//
// The environment holds references to the source the templates were created from.
// This makes it very inconvenient to pass around unless the templates are static
// strings.
//
// There are generally two ways to construct an environment:
//
//   - [NewEnvironment] creates an environment preconfigured with sensible
//     defaults.  It will contain all built-in filters, tests and globals as well
//     as a callback for auto escaping based on file extension.
//   - [NewEnvironmentEmpty] creates a completely blank environment.
type Environment struct {
	templates         templateStore
	filters           map[string]BoxedFilter
	tests             map[string]BoxedTest
	globals           map[string]Value
	defaultAutoEscape AutoEscapeFunc
	undefinedBehavior UndefinedBehavior
	formatter         formatterFunc
	debug             bool
}

// AutoEscapeFunc is the type of the function called by an Environment to
// determine the escaping behavior for the template of the specified name.
type AutoEscapeFunc func(name string) AutoEscape

type formatterFunc = func(*output, *State, Value) error

// NewEnvironment creates a new environment with sensible defaults.
//
// This environment does not yet contain any templates but it will have all
// the default filters, tests and globals loaded.  If you do not want any
// default configuration you can use the alternative
// [NewEnvironmentEmpty] method.
func NewEnvironment() *Environment {
	return &Environment{
		templates:         *newLoaderStoreDefault(),
		filters:           getDefaultBuiltinFilters(),
		tests:             getDefaultBuiltinTests(),
		globals:           getDefaultGlobals(),
		defaultAutoEscape: DefaultAutoEscapeCallback,
		formatter:         escapeFormatter,
	}
}

// NewEnvironmentEmpty creates a completely empty environment.
//
// This environment has no filters, no templates, no globals and no default
// logic for auto escaping configured.
func NewEnvironmentEmpty() *Environment {
	return &Environment{
		templates:         *newLoaderStoreDefault(),
		filters:           make(map[string]BoxedFilter),
		tests:             make(map[string]BoxedTest),
		globals:           make(map[string]Value),
		defaultAutoEscape: noAutoEscape,
		formatter:         escapeFormatter,
	}
}

// AddTemplate loads a template from a string into the environment.
//
// The `name` parameter defines the name of the template which identifies
// it.  To look up a loaded template use the [Environment.GetTemplate]
// method.
//
// Note that there are situations where the interface of this method is
// too restrictive as you need to hold on to the strings for the lifetime
// of the environment.
func (e *Environment) AddTemplate(name, source string) error {
	return e.templates.insert(name, source)
}

// SetLoader registers a template loader as source of templates.
//
// When a template loader is registered, the environment gains the ability
// to dynamically load templates.  The loader is invoked with the name of
// the template.  If this template exists the template source has to be returned,
// if the template does not exist an [Error] created with [NewErrorNotFound] has to
// be returned.  Once a template has been loaded it's stored on the environment.
// This means the loader is only invoked once per template name.
//
// For loading templates from the file system, you can use the
// [PathLoader] function.
func (e *Environment) SetLoader(f LoadFunc) {
	e.templates.setLoader(f)
}

// SetKeepTrailingNewline preserve the trailing newline when rendering templates.
//
// The default is `false`, which causes a single newline, if present, to be
// stripped from the end of the template.
func (e *Environment) SetKeepTrailingNewline(yes bool) {
	e.templates.KeepTrailingNewline = yes
}

// KeepTrailingNewline returns the value of the trailing newline preservation flag.
func (e *Environment) KeepTrailingNewline() bool {
	return e.templates.KeepTrailingNewline
}

// RemoveTemplate removes a template by name.
func (e *Environment) RemoveTemplate(name string) {
	e.templates.remove(name)
}

// ClearTemplates removes all stored templates.
//
// This method is mainly useful when combined with a loader as it causes
// the loader to "reload" templates.  By calling this method one can trigger
// a reload.
func (e *Environment) ClearTemplates() {
	e.templates.clear()
}

// GetTemplate fetches a template by name.
//
// This requires that the template has been loaded with
// [Environment.AddTemplate] beforehand.  If the template was
// not loaded an error of kind [TemplateNotFound] is returned.  If a loaded was
// added to the engine this can also dynamically load templates.
func (e *Environment) GetTemplate(name string) (*Template, error) {
	compiled, err := e.templates.get(name)
	if err != nil {
		return nil, err
	}
	if compiled == nil {
		return nil, NewErrorNotFound(name)
	}
	return &Template{
		env:               e,
		compiled:          compiled,
		initialAutoEscape: e.initialAutoEscape(name),
	}, nil
}

// TemplateFromNamedStr loads a template from a string.
//
// In some cases you really only need to work with (eg: render) a template to be
// rendered once only.
func (e *Environment) TemplateFromNamedStr(name, source string) (*Template, error) {
	compiled, err := newCompiledTemplate(name, source, *e.syntaxConfig(), e.KeepTrailingNewline())
	if err != nil {
		return nil, err
	}
	return &Template{
		env:               e,
		compiled:          compiled,
		initialAutoEscape: e.initialAutoEscape(name),
	}, nil
}

// TemplateFromStr loads a template from a string, with name `<string>`.
//
// This is a shortcut to [Environment.TemplateFromNamedStr]
// with name set to `<string>`.
func (e *Environment) TemplateFromStr(source string) (*Template, error) {
	return e.TemplateFromNamedStr("<string>", source)
}

// RenderNamedStr parses and renders a template from a string in one go with name.
//
// Like [Environment.RenderStr], but provide a name for the
// template to be used instead of the default `<string>`.  This is an
// alias for [Environment.TemplateFromNamedStr] paired with
// [Environment.Render].
func (e *Environment) RenderNamedStr(name, source string, ctx Value) (string, error) {
	tmpl, err := e.TemplateFromNamedStr(name, source)
	if err != nil {
		return "", err
	}
	return tmpl.Render(ctx)
}

// RenderStr parses and renders a template from a string in one go.
//
// In some cases you really only need a template to be rendered once from
// a string and returned.  The internal name of the template is `<string>`.
//
// This is an alias for [Environment.TemplateFromStr] paired with
// [Environment.Render].
func (e *Environment) RenderStr(source string, ctx Value) (string, error) {
	tmpl, err := e.TemplateFromStr(source)
	if err != nil {
		return "", err
	}
	return tmpl.Render(ctx)
}

// SetAutoEscapeCallback sets a new function to select the default auto escaping.
//
// This function is invoked when templates are loaded from the environment
// to determine the default auto escaping behavior.  The function is
// invoked with the name of the template and can make an initial auto
// escaping decision based on that.  The default implementation
// [DefaultAutoEscapeCallback].
// turn on escaping depending on the file extension.
func (e *Environment) SetAutoEscapeCallback(fn AutoEscapeFunc) {
	e.defaultAutoEscape = fn
}

// SetUndefinedBehavior changes the undefined behavior.
//
// This changes the runtime behavior of [Undefined] values in
// the template engine.  For more information see [UndefinedBehavior].  The
// default is [UndefinedBehaviorLenient].
func (e *Environment) SetUndefinedBehavior(behavior UndefinedBehavior) {
	e.undefinedBehavior = behavior
}

// UndefinedBehavior returns the current undefined behavior.
//
// This is particularly useful if a filter function or similar wants to change its
// behavior with regards to undefined values.
func (e *Environment) UndefinedBehavior() UndefinedBehavior {
	return e.undefinedBehavior
}

// SetDebug enables or disables the debug mode.
//
// When the debug mode is enabled the engine will dump out some of the
// execution state together with the source information of the executing
// template when an error is created.  The cost of this is relatively
// high as the data including the template source is cloned.
//
// When this is enabled templates will print debug information with source
// context when the error is printed.
func (e *Environment) SetDebug(enabled bool) {
	e.debug = enabled
}

// Debug returns the current value of the debug flag.
func (e *Environment) Debug() bool { return e.debug }

func (e *Environment) syntaxConfig() *syntaxConfig {
	return &e.templates.SyntaxConfig
}

// CompileExpression compiles an expression.
//
// This lets one compile an expression in the template language and
// receive the output.  This lets one use the expressions of the language
// be used as a minimal scripting language.  For more information and an
// example see [Expression].
func (e *Environment) CompileExpression(expr string) (*Expression, error) {
	ast, err := parseExpr(expr, *e.syntaxConfig())
	if err != nil {
		return nil, err
	}
	gen := newCodeGenerator("<expression>", expr)
	gen.compileExpr(ast)
	insts, _ := gen.finish()
	return newExpression(e, insts), nil
}

// AddFilter adds a new filter function.
func (e *Environment) AddFilter(name string, filter BoxedFilter) {
	e.filters[name] = filter
}

// RemoveFilter removes a filter by name.
func (e *Environment) RemoveFilter(name string) {
	delete(e.filters, name)
}

// AddTest adds a new test function.
//
// Test functions are similar to filters but perform a check on a value
// where the return value is always true or false.
func (e *Environment) AddTest(name string, test BoxedTest) {
	e.tests[name] = test
}

// RemoveTest removes a test by name.
func (e *Environment) RemoveTest(name string) {
	delete(e.tests, name)
}

// AddFunction adds a new global function.
func (e *Environment) AddFunction(name string, fn BoxedFunc) {
	e.globals[name] = valueFromBoxedFunc(fn)
}

// AddGlobal adds a new global variable.
func (e *Environment) AddGlobal(name string, val Value) {
	e.globals[name] = val
}

// RemoveGlobal a global function or variable by name.
func (e *Environment) RemoveGlobal(name string) {
	delete(e.globals, name)
}

func (e *Environment) format(v Value, state *State, out *output) error {
	if v.isUndefined() && e.undefinedBehavior == UndefinedBehaviorStrict {
		return NewError(UndefinedError, "")
	}
	return e.formatter(out, state, v)
}

func (e *Environment) getGlobal(name string) option.Option[Value] {
	val, ok := e.globals[name]
	if ok {
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
