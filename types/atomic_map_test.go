package types

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
)

func TestMap_BasicCRUD(t *testing.T) {
	m := NewAtomicMap[string, int]()

	// Put & GetWeight
	m.Put("apple", 10)
	m.Put("banana", 20)
	assert.Equal(t, 2, m.Len())

	// Get
	val, ok := m.Get("apple")
	assert.True(t, ok)
	assert.Equal(t, 10, val)

	_, ok = m.Get("cherry")
	assert.False(t, ok)

	// Exists
	assert.True(t, m.Exists("banana"))
	assert.False(t, m.Exists("durian"))

	// Remove
	m.Remove("apple")
	assert.Equal(t, 1, m.Len())
	assert.False(t, m.Exists("apple"))
}

func TestMap_PutIfNotExists(t *testing.T) {
	m := NewAtomicMap[string, int]()

	added := m.PutIfNotExists("apple", 100)
	assert.True(t, added)
	assert.Equal(t, 1, m.Len())

	added = m.PutIfNotExists("apple", 200)
	assert.False(t, added)
	val, _ := m.Get("apple")
	assert.Equal(t, 100, val)
}

func TestMap_Clear(t *testing.T) {
	m := NewAtomicMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)
	m.Clear()
	assert.Equal(t, 0, m.Len())
	assert.False(t, m.Exists("a"))
}

func TestMap_KeysValuesAndIterator(t *testing.T) {
	m := NewAtomicMap[string, int]()
	m.Put("a", 1)
	m.Put("b", 2)

	// Keys
	keys := m.Keys()
	assert.ElementsMatch(t, []string{"a", "b"}, keys)

	// Values
	values := m.Values()
	assert.ElementsMatch(t, []int{1, 2}, values)

	// Iterator
	count := 0
	for range m.Iterator() {
		count++
	}
	assert.Equal(t, 2, count)
}

func TestMap_RaceCondition(t *testing.T) {
	m := NewAtomicMap[string, int]()
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key_" + strconv.Itoa(i)

			// Put, PutIfNotExists
			m.Put(key, i)
			m.PutIfNotExists(key, i*100)

			// Get, Exists, GetWeight
			_, _ = m.Get(key)
			_ = m.Exists(key)
			_ = m.Len()

			// Remove
			m.Remove(key)
		}(i)
	}

	wg.Wait()

	assert.Equal(t, 0, m.Len())
}
