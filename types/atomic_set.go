package types

import "sync"

type AtomicSet[K comparable] struct {
	lock sync.RWMutex
	m    map[K]struct{}
}

func NewAtomicSet[K comparable]() *AtomicSet[K] {
	return &AtomicSet[K]{
		m: make(map[K]struct{}),
	}
}

func (s *AtomicSet[K]) Put(k K) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.m[k] = struct{}{}
}

func (s *AtomicSet[K]) Contains(k K) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, ok := s.m[k]
	return ok
}

func (s *AtomicSet[K]) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.m)
}

func (s *AtomicSet[K]) Remove(k K) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.m, k)
}

func (s *AtomicSet[K]) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.m = make(map[K]struct{})
}

func (s *AtomicSet[K]) Values() []K {
	s.lock.RLock()
	defer s.lock.RUnlock()

	keys := make([]K, 0, len(s.m))
	for k := range s.m {
		keys = append(keys, k)
	}
	return keys
}

func (s *AtomicSet[K]) Iterator() func(yield func(K) bool) {
	return func(yield func(K) bool) {
		s.lock.RLock()
		defer s.lock.RUnlock()
		for k := range s.m {
			if !yield(k) {
				return
			}
		}
	}
}
