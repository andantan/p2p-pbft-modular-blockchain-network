package atomic

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"sync"
)

type Cache[K comparable, V any] struct {
	lock    sync.RWMutex
	lookup  map[K]V
	ordered *types.List[K]
}

func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		lookup:  make(map[K]V),
		ordered: types.NewList[K](),
	}
}

func (c *Cache[K, V]) Add(k K, v V) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.lookup[k]; !ok {
		c.lookup[k] = v
		c.ordered.Insert(k)
	}
}

func (c *Cache[K, V]) Get(k K) (V, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	val, ok := c.lookup[k]
	return val, ok
}

func (c *Cache[K, V]) Remove(k K) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.lookup[k]; ok {
		delete(c.lookup, k)
	}

	if c.ordered.Contains(k) {
		c.ordered.Remove(k)
	}
}

func (c *Cache[K, V]) Prune(keysToRemove []K) {
	c.lock.Lock()
	defer c.lock.Unlock()

	toRemoveSet := make(map[K]struct{}, len(keysToRemove))

	for _, k := range keysToRemove {
		toRemoveSet[k] = struct{}{}
	}

	for k := range toRemoveSet {
		delete(c.lookup, k)
	}

	data := c.ordered.GetData()
	newOrderedKeys := make([]K, 0, len(data))

	for _, k := range data {
		if _, found := toRemoveSet[k]; !found {
			newOrderedKeys = append(newOrderedKeys, k)
		}
	}

	c.ordered.SetData(newOrderedKeys)
}

func (c *Cache[K, V]) First() (K, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.ordered.First()
}

func (c *Cache[K, V]) Count() int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return len(c.lookup)
}

func (c *Cache[K, V]) Contains(k K) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	_, ok := c.lookup[k]
	return ok
}

func (c *Cache[K, V]) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.lookup = make(map[K]V)
	c.ordered.Clear()
}

func (c *Cache[K, V]) GetLookup() map[K]V {
	return c.lookup
}

func (c *Cache[K, V]) Values() []V {
	c.lock.RLock()
	defer c.lock.RUnlock()

	orderedKeys := c.ordered.GetData()

	values := make([]V, len(orderedKeys))

	for i, k := range orderedKeys {
		values[i] = c.lookup[k]
	}

	return values
}

func (c *Cache[K, V]) LookupIterator() func(yield func(K, V) bool) {
	return func(yield func(K, V) bool) {
		c.lock.RLock()
		defer c.lock.RUnlock()

		for k, v := range c.lookup {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (c *Cache[K, V]) OrderedIterator() func(yield func(K) bool) {
	return func(yield func(K) bool) {
		c.lock.RLock()
		defer c.lock.RUnlock()

		listIterator := c.ordered.Iterator()
		listIterator(yield)
	}
}
