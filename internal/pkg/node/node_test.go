package node

import (
	"context"
	"testing"
	"time"
)

func ref[T any](v T) *T { return &v }

func TestNew(t *testing.T) {
	type in[T any] struct {
		v         *T
		completed bool
	}

	type out[T any] struct {
		set      bool
		value    T
		finished bool
	}

	type values[T any] struct {
		name string
		in   in[T]
		out  out[T]
	}

	for _, tt := range []values[int]{
		{
			name: "unset not completed",
			in:   in[int]{},
			out:  out[int]{},
		},
		{
			name: "set not completed",
			in: in[int]{
				v: ref(1),
			},
			out: out[int]{
				set:   true,
				value: 1,
			},
		},
		{
			name: "unset completed",
			in: in[int]{
				completed: true,
			},
			out: out[int]{
				finished: true,
			},
		},
		{
			name: "set completed",
			in: in[int]{
				v:         ref(1),
				completed: true,
			},
			out: out[int]{
				finished: true,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			n := New[int](tt.in.v, tt.in.completed)
			if tt.out.set {
				if n.set {
					if tt.out.value != n.value {
						t.Errorf("expected set value %d, got %d", tt.out.value, n.value)
					}
				} else {
					t.Errorf("value should be set")
				}
			} else {
				if n.set {
					t.Errorf("value should not be set")
				}
			}

			if n.next != nil {
				t.Errorf("next is not nil")
			}

			select {
			case <-n.wait:
				t.Errorf("node wait closed")
			default:
			}

			if tt.out.finished {
				if !n.finished {
					t.Errorf("node should be finished")
				}
				select {
				case <-n.completed:
				default:
					t.Errorf("node should be completed")
				}
			} else {
				if n.finished {
					t.Errorf("node should not be finished")
				}
				select {
				case <-n.completed:
					t.Errorf("node should not be completed")
				default:
				}
			}
		})
	}
}

func TestNode_SetNext(t *testing.T) {
	n1 := New[int](ref(1), false)
	n2 := New[int](ref(2), false)
	n1.SetNext(n2)

	select {
	case <-n1.wait:
	default:
		t.Errorf("wait should be closed")
	}

	if n1.next != n2 {
		t.Errorf("next is not set")
	}
}

func TestNode_Next(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	n := (*Node[int]).Next(nil, ctx)
	if n != nil {
		t.Errorf("expected nil next from nil node")
	}

	n1 := New[int](ref(1), false)
	n = n1.Next(ctx)
	if n != nil {
		t.Errorf("expected nil next from cancelled context")
	}

	n2 := New[int](nil, true)
	n1.SetNext(n2)

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	n = n2.Next(ctx)
	if n != nil {
		t.Errorf("expected nil next from completed node")
	}

	n = n1.Next(ctx)
	if n != n2 {
		t.Errorf("did not get expected next from node")
	}

	n1.finished = true
	n = n1.Next(ctx)
	if n != nil {
		t.Errorf("finished but not completed should return nil node")
	}
}

func TestGetCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if GetCtx(nil) != context.Background() {
		t.Errorf("did not get background context from nil")
	}
	if GetCtx(ctx) != ctx {
		t.Errorf("did not return same ctx from valid")
	}
}

func TestNode_SendValue(t *testing.T) {
	var n *Node[int]
	noValue := func(ctx context.Context) {
		c := make(chan int, 1)
		n.SendValue(c, ctx)
		select {
		case <-c:
			t.Errorf("got value from node")
		default:
		}
	}
	value := func(ctx context.Context, i int) {
		c := make(chan int, 1)
		n.SendValue(c, ctx)
		select {
		case v := <-c:
			if v != i {
				t.Errorf("got value %d from node, expected %d", v, i)
			}
		default:
			t.Errorf("did not get value from node")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n = nil
	noValue(ctx)

	n = New[int](nil, false)
	noValue(ctx)

	n = New[int](ref(1), false)
	value(ctx, 1)

	n = New[int](ref(1), false)
	ctx, cancel = context.WithTimeout(ctx, time.Second)
	defer cancel()
	c := make(chan int)
	n.SendValue(c, ctx)
	select {
	case <-c:
		t.Errorf("got value from node")
	default:
	}
}
