package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	s := NewStack(128)

	s.Push(1)
	s.Push(2)

	value := s.Pop()

	assert.Equal(t, value, 1)
	
	fmt.Println(s.data)
}

func TestStackBytes(t *testing.T) {
	s := NewStack(128)

	s.Push('a')
	s.Push('b')

	value := s.Pop()

	assert.Equal(t, value, 'a')
	
	fmt.Println(s.data)
}

func TestVM(t *testing.T) {
	data := []byte{0x02, 0x0a, 'a', 0x0c, 'a', 0x0c, 0x0d}
	vm := NewVM(data)

	assert.Nil(t, vm.Run())
	result := vm.stack.Pop().([]byte)
	assert.Equal(t, "aa", string(result))
}


