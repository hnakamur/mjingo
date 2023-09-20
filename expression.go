package mjingo

import "errors"

// Expression represents a compiled expression.
//
// An expression is created via the
// [Environment.CompileExpression] method.  It provides
// a method to evaluate the expression and return the result as value object.
// This for instance can be used to evaluate simple expressions from user
// provided input to implement features such as dynamic filtering.
type Expression struct {
	env   *Environment
	insts instructions
}

func newExpression(env *Environment, insts instructions) *Expression {
	return &Expression{env: env, insts: insts}
}

// Eval evaluates the expression with some context value.
//
// The result of the expression is returned as [Value].
func (e *Expression) Eval(root Value) (Value, error) {
	vm := newVirtualMachine(e.env)
	optVal, _, err := vm.eval(e.insts, root, make(map[string]instructions), newOutputNull(), autoEscapeNone{})
	if err != nil {
		return Value{}, err
	}
	if optVal.IsNone() {
		return Value{}, errors.New("expression evaluation did not leave value on stack")
	}
	return optVal.Unwrap(), nil
}
