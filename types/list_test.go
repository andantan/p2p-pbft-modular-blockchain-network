package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestList_InsertAndLen(t *testing.T) {
	l := NewList[int]()
	assert.Equal(t, 0, l.Len())

	l.Insert(10)
	l.Insert(20)

	assert.Equal(t, 2, l.Len())
}

func TestList_Get(t *testing.T) {
	l := NewList[string]()
	l.Insert("apple")
	l.Insert("banana")

	val, err := l.Get(1)
	assert.NoError(t, err)
	assert.Equal(t, "banana", val)

	_, err = l.Get(5)
	assert.Error(t, err)
}

func TestList_Pop(t *testing.T) {
	l := NewList[int]()
	l.Insert(1)
	l.Insert(2)
	l.Insert(3)

	l.Pop(1) // Remove '2'

	assert.Equal(t, 2, l.Len())
	val, _ := l.Get(0)
	assert.Equal(t, 1, val)
	val, _ = l.Get(1)
	assert.Equal(t, 3, val)
}

func TestList_Remove(t *testing.T) {
	l := NewList[string]()
	l.Insert("cat")
	l.Insert("dog")
	l.Insert("lion")

	l.Remove("dog")
	assert.Equal(t, 2, l.Len())
	assert.False(t, l.Contains("dog"))

	l.Remove("tiger")
	assert.Equal(t, 2, l.Len())
}

func TestList_Clear(t *testing.T) {
	l := NewList[int]()
	l.Insert(100)
	l.Insert(200)
	l.Clear()

	assert.Equal(t, 0, l.Len())
	assert.Equal(t, 0, len(l.data))
}

func TestList_GetIndexAndContains(t *testing.T) {
	l := NewList[rune]()
	l.Insert('a')
	l.Insert('b')

	assert.True(t, l.Contains('a'))
	assert.False(t, l.Contains('c'))

	idx, err := l.GetIndex('b')
	assert.NoError(t, err)
	assert.Equal(t, 1, idx)

	_, err = l.GetIndex('z')
	assert.Error(t, err)
}

func TestList_FirstAndLast(t *testing.T) {
	l := NewList[int]()

	_, err := l.First()
	assert.Error(t, err)
	_, err = l.Last()
	assert.Error(t, err)

	l.Insert(10)
	l.Insert(20)
	l.Insert(30)

	first, _ := l.First()
	assert.Equal(t, 10, first)
	last, _ := l.Last()
	assert.Equal(t, 30, last)
}

func TestList_GetData(t *testing.T) {
	l := NewList[int]()
	l.Insert(1)
	l.Insert(2)

	snapshot := l.GetData()
	assert.Equal(t, []int{1, 2}, snapshot)

	snapshot[0] = 99

	originalFirst, _ := l.First()
	assert.Equal(t, 1, originalFirst)
}

func TestList_SetData(t *testing.T) {
	l := NewList[int]()
	newData := []int{5, 6, 7}
	l.SetData(newData)

	assert.Equal(t, 3, l.Len())
	first, _ := l.First()
	assert.Equal(t, 5, first)
}

func TestList_Iterator(t *testing.T) {
	l := NewList[rune]()
	l.Insert('x')
	l.Insert('y')
	l.Insert('z')

	var results []rune
	for val := range l.Iterator() {
		results = append(results, val)
	}

	assert.Equal(t, []rune{'x', 'y', 'z'}, results)

	var count int
	for val := range l.Iterator() {
		count++
		if val == 'y' {
			break
		}
	}
	assert.Equal(t, 2, count)
}
