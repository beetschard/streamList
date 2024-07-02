package streamList

import (
	"context"
	"github.com/beetschard/streamList/internal/pkg/node"
	"sync"
)

type List[T any] struct {
	lock sync.Mutex
	head *node.Node[T]
	tail *node.Node[T]
}

func (l *List[T]) Stream(ctx context.Context) <-chan T {
	ctx = node.GetCtx(ctx)
	c := make(chan T)
	go func() {
		defer close(c)
		for n := l.getHead(); n != nil; n = n.Next(ctx) {
			n.SendValue(c, ctx)
		}
	}()
	return c
}

func (l *List[T]) Complete() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.append(node.New[T](nil, true), false)
}

func (l *List[T]) Append(value ...T) {
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, v := range value {
		l.append(node.New[T](&v, false), false)
	}
}

func (l *List[T]) Reset() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.reset()
}

func (l *List[T]) reset() {
	l.append(node.New[T](nil, false), true)
}

func (l *List[T]) append(next *node.Node[T], replaceHead bool) {
	if l.head == nil || replaceHead {
		l.head = next
	}
	if l.tail != nil {
		defer l.tail.SetNext(next)
	}
	l.tail = next
}

func (l *List[T]) getHead() *node.Node[T] {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.head == nil {
		l.reset()
	}
	return l.head
}
