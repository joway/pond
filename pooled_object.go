package pond

import (
	"time"
)

type pooledObject struct {
	object   interface{}
	returnAt time.Time
}

func newPooledObject(object interface{}) *pooledObject {
	return &pooledObject{
		object:   object,
		returnAt: time.Now(),
	}
}

func (o pooledObject) Object() interface{} {
	return o.object
}

func (o pooledObject) IdleTime() time.Duration {
	return time.Since(o.returnAt)
}

func (o *pooledObject) Returned() {
	o.returnAt = time.Now()
}
