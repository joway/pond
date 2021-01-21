package pond

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPooledObject(t *testing.T) {
	obj := &testObject{name: "test"}
	po := newPooledObject(obj)
	beg := po.returnAt
	time.Sleep(time.Millisecond * 100)
	idleTime := po.IdleTime()
	po.Returned()
	end := po.returnAt
	assert.Equal(t, end.Sub(beg).Milliseconds(), idleTime.Milliseconds())
}
