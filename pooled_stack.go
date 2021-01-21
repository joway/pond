package pond

//LIFO
type pooledStack struct {
	stack []*pooledObject
}

func newPooledStack() *pooledStack {
	return &pooledStack{
		stack: make([]*pooledObject, 0),
	}
}

func (p pooledStack) Len() int {
	return len(p.stack)
}

func (p *pooledStack) Push(po *pooledObject) {
	p.stack = append(p.stack, po)
}

//Pop pop top item
func (p *pooledStack) Pop() *pooledObject {
	n := p.Len() - 1
	if n < 0 {
		return nil
	}
	po := p.stack[n]
	p.stack = p.stack[:n]
	return po
}

func (p *pooledStack) Top() *pooledObject {
	n := p.Len() - 1
	if n < 0 {
		return nil
	}
	return p.stack[n]
}

//BPop pop bottom item
func (p *pooledStack) BPop() *pooledObject {
	if p.Len() <= 0 {
		return nil
	}
	po := p.stack[0]
	p.stack = p.stack[1:]
	return po
}

func (p pooledStack) Bottom() *pooledObject {
	if p.Len() == 0 {
		return nil
	}
	return p.stack[0]
}

func (p pooledStack) Range(handler func(object *pooledObject)) {
	for _, o := range p.stack {
		handler(o)
	}
}
