package types

import "sync"

type AtomicMap[K comparable, V any] struct {
	lock sync.RWMutex
	m    map[K]V
}

func NewAtomicMap[K comparable, V any]() *AtomicMap[K, V] {
	return &AtomicMap[K, V]{
		m: make(map[K]V),
	}
}

func (m *AtomicMap[K, V]) Get(k K) (V, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	val, ok := m.m[k]
	return val, ok
}

func (m *AtomicMap[K, V]) Put(k K, v V) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.m[k] = v
}

func (m *AtomicMap[K, V]) Exists(k K) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, ok := m.m[k]
	return ok
}

func (m *AtomicMap[K, V]) PutIfNotExists(k K, v V) bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.m[k]; ok {
		return false
	}

	m.m[k] = v

	return true
}

func (m *AtomicMap[K, V]) Len() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.m)
}

func (m *AtomicMap[K, V]) Remove(k K) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.m, k)
}

func (m *AtomicMap[K, V]) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.m = make(map[K]V)
}

func (m *AtomicMap[K, V]) Keys() []K {
	m.lock.RLock()
	defer m.lock.RUnlock()

	keys := make([]K, 0, len(m.m))
	for k := range m.m {
		keys = append(keys, k)
	}

	return keys
}

func (m *AtomicMap[K, V]) Values() []V {
	m.lock.RLock()
	defer m.lock.RUnlock()

	values := make([]V, 0, len(m.m))
	for _, v := range m.m {
		values = append(values, v)
	}

	return values
}

func (m *AtomicMap[K, V]) Iterator() func(yield func(K, V) bool) {
	return func(yield func(K, V) bool) {
		m.lock.RLock()
		defer m.lock.RUnlock()

		for k, v := range m.m {
			if !yield(k, v) {
				return
			}
		}
	}
}
