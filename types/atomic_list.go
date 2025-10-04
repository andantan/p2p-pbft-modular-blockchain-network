package types

import (
	"sync"
)

type AtomicList[T any] struct {
	lock sync.RWMutex
	list *List[T]
}

func NewAtomicList[T any]() *AtomicList[T] {
	return &AtomicList[T]{
		list: NewList[T](),
	}
}

func (l *AtomicList[T]) Insert(e T) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.list.Insert(e)
}

func (l *AtomicList[T]) Get(index int) (T, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.Get(index)
}

func (l *AtomicList[T]) Pop(index int) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.list.Pop(index)
}

func (l *AtomicList[T]) Remove(e T) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.list.Remove(e)
}

func (l *AtomicList[T]) Clear() {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.list.Clear()
}

func (l *AtomicList[T]) GetIndex(e T) (int, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.GetIndex(e)
}

func (l *AtomicList[T]) Contains(e T) bool {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.Contains(e)
}

func (l *AtomicList[T]) First() (T, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.First()
}

func (l *AtomicList[T]) Last() (T, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.Last()
}

func (l *AtomicList[T]) Len() int {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.Len()
}

func (l *AtomicList[T]) GetData() []T {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.GetData()
}

func (l *AtomicList[T]) Iterator() func(yield func(T) bool) {
	return func(yield func(T) bool) {
		l.lock.RLock()
		defer l.lock.RUnlock()

		listIterator := l.list.Iterator()
		listIterator(yield)
	}
}
