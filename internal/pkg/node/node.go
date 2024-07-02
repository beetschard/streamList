package node

import "context"

type Node[T any] struct {
	wait      chan struct{}
	completed chan struct{}
	next      *Node[T]
	set       bool
	finished  bool
	value     T
}

func New[T any](value *T, completed bool) *Node[T] {
	set := (value != nil) && !completed
	n := &Node[T]{
		wait:      make(chan struct{}),
		completed: make(chan struct{}),
		set:       set,
		finished:  completed,
	}

	if set {
		n.value = *value
	}

	if completed {
		close(n.completed)
	}

	return n
}

func (n *Node[T]) SetNext(next *Node[T]) {
	n.next = next
	close(n.wait)
}

func (n *Node[T]) Next(ctx context.Context) *Node[T] {
	if n == nil {
		return nil
	}
	return n.nextWait(GetCtx(ctx))
}

func (n *Node[T]) nextWait(ctx context.Context) *Node[T] {
	select {
	case <-ctx.Done():
	case <-n.completed:
	case <-n.wait:
		if !n.finished {
			return n.next
		}
	}
	return nil
}

func (n *Node[T]) SendValue(c chan<- T, ctx context.Context) {
	if n != nil && n.set {
		n.sendValueWait(c, GetCtx(ctx))
	}
}

func (n *Node[T]) sendValueWait(c chan<- T, ctx context.Context) {
	select {
	case <-ctx.Done():
	case c <- n.value:
	}
}

func GetCtx(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
