package mjingo

import "errors"

type Expression struct {
	env   *Environment
	insts instructions
}

func newExpression(env *Environment, insts instructions) *Expression {
	return &Expression{env: env, insts: insts}
}

func (e *Expression) Eval(root Value) (Value, error) {
	vm := newVirtualMachine(e.env)
	optVal, err := vm.eval(e.insts, root, make(map[string]instructions), newOutputNull(), autoEscapeNone{})
	if err != nil {
		return nil, err
	}
	if optVal.IsNone() {
		return nil, errors.New("expression evaluation did not leave value on stack")
	}
	return optVal.Unwrap(), nil
}
