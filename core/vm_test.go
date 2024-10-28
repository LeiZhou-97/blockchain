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
	// F O O => pack [F O O]
	data := []byte{0x03, 0x0a, 'F', 0x0c, 'O', 0x0c, 'O', 0x0c, 0x0d,  0x05, 0x0a, 0x0f}
	// data := []byte{0x02, 0x0a, 'a', 0x0c, 'a', 0x0c, 0x0d}
	contractState := NewState()
	vm := NewVM(data, contractState)

	assert.Nil(t, vm.Run())
	valueBytes, err := contractState.Get([]byte("FOO"))
	value := deserializeInt64(valueBytes)
	assert.Nil(t, err)
	assert.Equal(t, value, int64(5))
	//assert.Equal(t, "aa", string(result))
}


