package pond

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestPoolManager(t *testing.T) {
	pm := newPoolManager()
	assert.Equal(t, 0, pm.ActiveSize())
	assert.Equal(t, 0, pm.IdleSize())

	//create by order
	loopSize := 100
	for i := 0; i < loopSize; i++ {
		pm.Create(&testObject{name: strconv.Itoa(i)})
		time.Sleep(time.Millisecond)
	}
	assert.Equal(t, 0, pm.ActiveSize())
	assert.Equal(t, loopSize, pm.IdleSize())

	//borrow and return
	for i := 0; i < loopSize; i++ {
		pobj := pm.Borrow()
		obj := pobj.Object().(*testObject)
		assert.Equal(t, strconv.Itoa(loopSize-1), obj.name)
		assert.Equal(t, 1, pm.ActiveSize())
		pm.Return(obj)
	}
	assert.Equal(t, 0, pm.ActiveSize())
	assert.Equal(t, loopSize, pm.IdleSize())

	//range
	count := 0
	pm.Range(func(object interface{}) {
		count++
	})
	assert.Equal(t, loopSize, count)

	//deactivate
	for i := 1; i <= loopSize; i++ {
		pobj := pm.Borrow()
		assert.Equal(t, 1, pm.ActiveSize())
		pm.Deactivate(pobj.Object())
		assert.Equal(t, 0, pm.ActiveSize())
		assert.Equal(t, loopSize-i, pm.IdleSize())
	}
	assert.Equal(t, 0, pm.ActiveSize())
	assert.Equal(t, 0, pm.IdleSize())
}
