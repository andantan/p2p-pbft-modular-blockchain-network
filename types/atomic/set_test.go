package atomic

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
)

func TestSet_BasicOperations(t *testing.T) {
	s := NewSet[string]()

	// Put & Len
	s.Put("apple")
	s.Put("banana")
	assert.Equal(t, 2, s.Len())
	s.Put("apple")
	assert.Equal(t, 2, s.Len())

	// Contains
	assert.True(t, s.Contains("banana"))
	assert.False(t, s.Contains("cherry"))

	// Remove
	s.Remove("apple")
	assert.Equal(t, 1, s.Len())
	assert.False(t, s.Contains("apple"))
}

func TestSet_Clear(t *testing.T) {
	s := NewSet[int]()
	s.Put(1)
	s.Put(2)
	s.Clear()
	assert.Equal(t, 0, s.Len())
	assert.False(t, s.Contains(1))
}

func TestSet_ValuesAndIterator(t *testing.T) {
	s := NewSet[string]()
	s.Put("a")
	s.Put("b")

	// Values
	values := s.Values()
	assert.ElementsMatch(t, []string{"a", "b"}, values)

	// Iterator
	count := 0
	for range s.Iterator() {
		count++
	}
	assert.Equal(t, 2, count)
}

func TestSet_RaceCondition(t *testing.T) {
	s := NewSet[string]()
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key_" + strconv.Itoa(i)

			s.Put(key)
			_ = s.Contains(key)
			_ = s.Len()
			s.Remove(key)
		}(i)
	}

	wg.Wait()

	assert.Equal(t, 0, s.Len(), "모든 키가 추가된 후 다시 제거되어야 합니다.")
}
