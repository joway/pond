package pond

//poolManager is not thread-safe
type poolManager struct {
	idle   *pooledStack
	active map[interface{}]*pooledObject
}

func newPoolManager() *poolManager {
	return &poolManager{
		idle:   newPooledStack(),
		active: make(map[interface{}]*pooledObject),
	}
}

func (p *poolManager) Earliest() *pooledObject {
	return p.idle.Bottom()
}

func (p *poolManager) PopEarliest() *pooledObject {
	return p.idle.BPop()
}

func (p *poolManager) Latest() *pooledObject {
	return p.idle.Top()
}

func (p *poolManager) PopLatest() *pooledObject {
	return p.idle.Pop()
}

func (p *poolManager) Borrow() *pooledObject {
	po := p.PopLatest()
	if po == nil {
		return nil
	}
	p.active[po.Object()] = po
	return po
}

func (p *poolManager) Create(object interface{}) {
	po, existed := p.active[object]
	//object is nil or existed in active
	if object == nil || existed {
		return
	}
	//create new one
	po = newPooledObject(object)
	p.idle.Push(po)
}

func (p *poolManager) Return(object interface{}) {
	po := p.active[object]
	if po == nil {
		//return a object that not existed
		return
	}
	delete(p.active, object)
	po.Returned()
	p.idle.Push(po)
}

func (p *poolManager) Deactivate(object interface{}) {
	delete(p.active, object)
}

func (p *poolManager) ActiveSize() int {
	return len(p.active)
}

func (p *poolManager) IdleSize() int {
	return p.idle.Len()
}

func (p *poolManager) Size() int {
	return p.ActiveSize() + p.IdleSize()
}

func (p *poolManager) RangeIdle(fn func(object interface{})) {
	p.idle.Range(func(po *pooledObject) {
		fn(po.Object())
	})
}

func (p *poolManager) RangeActive(fn func(object interface{})) {
	for _, po := range p.active {
		fn(po.Object())
	}
}

func (p *poolManager) Range(fn func(object interface{})) {
	p.RangeIdle(fn)
	p.RangeActive(fn)
}
