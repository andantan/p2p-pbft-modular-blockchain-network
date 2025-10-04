package atomic

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestList_ConcurrentInsert(t *testing.T) {
	list := NewList[int]()
	var wg sync.WaitGroup
	numGoroutines := 100
	numInsertsPerGoroutine := 10

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numInsertsPerGoroutine; j++ {
				list.Insert(j)
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, numGoroutines*numInsertsPerGoroutine, list.Len())
}

func TestList_ConcurrentReadWrite(t *testing.T) {
	list := NewList[int]()
	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < 100; i++ {
		list.Insert(i)
	}

	wg.Add(numGoroutines * 3)

	// read goroutine
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_, _ = list.Get(10)
			_ = list.Len()
			_ = list.Contains(20)
		}()
	}

	// write goroutine
	for i := 0; i < numGoroutines; i++ {
		go func(val int) {
			defer wg.Done()
			list.Insert(100 + val)
		}(i)
	}

	// delete goroutine
	for i := 0; i < numGoroutines; i++ {
		go func(val int) {
			defer wg.Done()
			list.Remove(val)
		}(i)
	}

	wg.Wait()

	assert.Equal(t, 100, list.Len())
}

func TestList_IteratorConcurrent(t *testing.T) {
	list := NewList[int]()
	for i := 0; i < 1000; i++ {
		list.Insert(i)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// iterator goroutine
	go func() {
		defer wg.Done()
		count := 0
		for range list.Iterator() {
			count++
		}
		assert.Equal(t, 999, count)
	}()

	// remove goroutine
	go func() {
		defer wg.Done()
		list.Remove(500)
	}()

	wg.Wait()

	assert.Equal(t, 999, list.Len())
}
