package streamList

import (
	"context"
	"iter"
	"sync"
)

type (
	List[T any] struct {
		lock sync.Mutex
		head *node[T]
		tail *node[T]
	}
	node[T any] struct {
		wait  chan struct{}
		next  *node[T]
		set   bool
		value T
	}
)

func (l *List[T]) Reset() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.reset()
}

func (l *List[T]) Append(value ...T) {
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, v := range value {
		l.append(newNode[T](&v), false)
	}
}

func (l *List[T]) Iter(ctx context.Context) iter.Seq[T] {
	return func(yield func(T) bool) {
		ctx = getCtx(ctx)
		for n := l.getHead(); ctx.Err() == nil; n = n.next {
			if n.set {
				if !yield(n.value) {
					return
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-n.wait:
			}
		}
	}
}

func getCtx(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func (l *List[T]) getHead() *node[T] {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.head == nil {
		l.reset()
	}
	return l.head
}

func (l *List[T]) reset() {
	l.append(newNode[T](nil), true)
}

func (l *List[T]) append(next *node[T], replaceHead bool) {
	if l.head == nil || replaceHead {
		l.head = next
	}
	if l.tail != nil {
		l.tail.next = next
		defer close(l.tail.wait)
	}
	l.tail = next
}

func newNode[T any](value *T) *node[T] {
	set := value != nil
	n := &node[T]{
		wait: make(chan struct{}),
		set:  set,
	}

	if set {
		n.value = *value
	}

	return n
}
