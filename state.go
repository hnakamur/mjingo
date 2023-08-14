package mjingo

type virtualMachineState struct {
	env          *Environment
	ctx          context
	currentBlock option[string]
	instructions instructions
	blocks       map[string]blockStack
}

type locals = map[string]value

type blockStack struct {
	instrs []instructions
	depth  uint
}

func (s *virtualMachineState) lookup(name string) option[value] {
	return s.ctx.load(s.env, name)
}

func newBlockStack(instrs instructions) blockStack {
	return blockStack{instrs: []instructions{instrs}, depth: 0}
}

func (b *blockStack) instructions() instructions {
	return b.instrs[b.depth]
}

func (b *blockStack) push() bool {
	if b.depth+1 < uint(len(b.instrs)) {
		b.depth++
		return true
	}
	return false
}

func (b *blockStack) pop() {
	if b.depth > 0 {
		b.depth--
	}
}

func (b *blockStack) appendInstructions(instrs instructions) {
	b.instrs = append(b.instrs, instrs)
}
