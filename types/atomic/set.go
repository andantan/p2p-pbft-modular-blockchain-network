package atomic

import "sync"

type Set[K comparable] struct {
	lock sync.RWMutex
	m    map[K]struct{}
}

func NewSet[K comparable]() *Set[K] {
	return &Set[K]{
		m: make(map[K]struct{}),
	}
}

func (s *Set[K]) Put(k K) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.m[k] = struct{}{}
}

func (s *Set[K]) Contains(k K) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, ok := s.m[k]
	return ok
}

func (s *Set[K]) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.m)
}

func (s *Set[K]) Remove(k K) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.m, k)
}

func (s *Set[K]) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.m = make(map[K]struct{})
}

func (s *Set[K]) Values() []K {
	s.lock.RLock()
	defer s.lock.RUnlock()

	keys := make([]K, 0, len(s.m))
	for k := range s.m {
		keys = append(keys, k)
	}
	return keys
}

func (s *Set[K]) Iterator() func(yield func(K) bool) {
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
