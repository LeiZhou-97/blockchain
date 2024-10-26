package core


type Instruction byte

const (
	InstrPushInt Instruction = 0x0a
	InstrAdd Instruction = 0x0b
	InstrPushByte Instruction = 0x0c
	InstrPushPack Instruction = 0x0d
	InstrSub Instruction = 0x0e
)

type Stack struct {
	data []any //interface{}
	sp int
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]any, size),
		sp: 0,
	}
}

func (s *Stack) Push(v any) {
	s.data[s.sp] = v
	s.sp++
}

func (s *Stack) Pop() any {
	value := s.data[0]
	copy(s.data, s.data[1:])
	s.sp--

	return value
}

type VM struct {
	data []byte
	ip int // instruction pointer
	stack *Stack
	sp int // stack pointer
}

func NewVM(data []byte) *VM {
	return &VM{
		data: data,
		ip: 0,
		stack: NewStack(128),
		sp: -1,
	}
}

func (vm *VM) Run() error {
	for {
		instr := Instruction(vm.data[vm.ip])


		if err := vm.Exec(instr); err != nil {
			return err
		}

		vm.ip++

		if vm.ip > len(vm.data) - 1 {
			break
		}
	}

	return nil
}

func (vm *VM) Exec(instr Instruction) error {
	switch instr {
	case InstrPushInt:
		vm.stack.Push(int(vm.data[vm.ip-1]))
	case InstrPushByte:
		vm.stack.Push(byte(vm.data[vm.ip-1]))
	case InstrPushPack:
		n := vm.stack.Pop().(int)
		b := make([]byte, n)
		for i:=0; i<n; i++ {
			b[i] = vm.stack.Pop().(byte)
		}
		vm.stack.Push(b)
	case InstrAdd:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a + b
		vm.stack.Push(c)
	case InstrSub:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a - b
		vm.stack.Push(c)
	}

	return nil
}

