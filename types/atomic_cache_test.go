package types

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
)

func TestCache_BasicCRUD(t *testing.T) {
	c := NewAtomicCache[string, int]()

	// Put & Count
	c.Put("apple", 10)
	c.Put("banana", 20)
	assert.Equal(t, 2, c.Count())
	c.Put("apple", 100)
	assert.Equal(t, 2, c.Count())

	// Get
	val, ok := c.Get("apple")
	assert.True(t, ok)
	assert.Equal(t, 10, val)

	// Contains
	assert.True(t, c.Contains("banana"))
	assert.False(t, c.Contains("cherry"))

	// Remove
	c.Remove("apple")
	assert.Equal(t, 1, c.Count())
	assert.False(t, c.Contains("apple"))
}

func TestCache_Prune(t *testing.T) {
	c := NewAtomicCache[string, int]()
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)
	c.Put("d", 4)

	c.Prune([]string{"b", "d", "e"})

	assert.Equal(t, 2, c.Count())
	assert.True(t, c.Contains("a"))
	assert.False(t, c.Contains("b"))
	assert.True(t, c.Contains("c"))
	assert.False(t, c.Contains("d"))
}

func TestCache_First(t *testing.T) {
	c := NewAtomicCache[string, string]()
	c.Put("second", "B")
	c.Put("first", "A")

	firstKey, err := c.First()
	assert.NoError(t, err)
	assert.Equal(t, "second", firstKey)
}

func TestCache_Clear(t *testing.T) {
	c := NewAtomicCache[string, int]()
	c.Put("a", 1)
	c.Clear()
	assert.Equal(t, 0, c.Count())
	assert.False(t, c.Contains("a"))
}

func TestCache_ValuesAndIterators(t *testing.T) {
	c := NewAtomicCache[string, int]()
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)

	// Values
	values := c.Values()
	assert.Equal(t, []int{1, 2, 3}, values)

	var orderedKeys []string
	for k := range c.OrderedIterator() {
		orderedKeys = append(orderedKeys, k)
	}
	assert.Equal(t, []string{"a", "b", "c"}, orderedKeys)

	count := 0
	for range c.LookupIterator() {
		count++
	}
	assert.Equal(t, 3, count)
}

func TestCache_RaceCondition(t *testing.T) {
	c := NewAtomicCache[string, int]()
	var wg sync.WaitGroup
	numGoroutines := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			key := "key_" + strconv.Itoa(i)

			c.Put(key, i)
			_, _ = c.Get(key)
			_ = c.Contains(key)
			c.Remove(key)
		}(i)
	}

	wg.Wait()
	assert.Equal(t, 0, c.Count())
}
