package atomic

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"sync"
)

type List[T any] struct {
	lock sync.RWMutex
	list *types.List[T]
}

func NewList[T any]() *List[T] {
	return &List[T]{
		list: types.NewList[T](),
	}
}

func (l *List[T]) Insert(e T) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.list.Insert(e)
}

func (l *List[T]) Get(index int) (T, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.Get(index)
}

func (l *List[T]) Pop(index int) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.list.Pop(index)
}

func (l *List[T]) Remove(e T) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.list.Remove(e)
}

func (l *List[T]) Clear() {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.list.Clear()
}

func (l *List[T]) GetIndex(e T) (int, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.GetIndex(e)
}

func (l *List[T]) Contains(e T) bool {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.Contains(e)
}

func (l *List[T]) First() (T, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.First()
}

func (l *List[T]) Last() (T, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.Last()
}

func (l *List[T]) Len() int {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.Len()
}

func (l *List[T]) GetData() []T {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.GetData()
}

func (l *List[T]) Iterator() func(yield func(T) bool) {
	return func(yield func(T) bool) {
		l.lock.RLock()
		defer l.lock.RUnlock()

		listIterator := l.list.Iterator()
		listIterator(yield)
	}
}
