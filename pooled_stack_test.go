package pond

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestPooledStack(t *testing.T) {
	stk := newPooledStack()
	size := 100
	assert.Equal(t, 0, stk.Len())

	//test pop
	for i := 0; i < size; i++ {
		stk.Push(newPooledObject(&testObject{name: strconv.Itoa(i)}))
	}
	assert.Equal(t, size, stk.Len())
	for i := 0; i < size; i++ {
		top := stk.Top()
		pop := stk.Pop()
		assert.Equal(t, top, pop)
		n := size - i - 1
		assert.Equal(t, strconv.Itoa(n), pop.Object().(*testObject).name)
		assert.Equal(t, n, stk.Len())
	}
	assert.Equal(t, 0, stk.Len())
	assert.Nil(t, stk.Pop())

	//test bpop
	for i := 0; i < size; i++ {
		stk.Push(newPooledObject(&testObject{name: strconv.Itoa(i)}))
	}
	assert.Equal(t, size, stk.Len())
	for i := 0; i < size; i++ {
		bottom := stk.Bottom()
		pop := stk.BPop()
		assert.Equal(t, bottom, pop)
		assert.Equal(t, strconv.Itoa(i), pop.Object().(*testObject).name)
		assert.Equal(t, size-i-1, stk.Len())
	}
	assert.Equal(t, 0, stk.Len())
	assert.Nil(t, stk.BPop())
}
