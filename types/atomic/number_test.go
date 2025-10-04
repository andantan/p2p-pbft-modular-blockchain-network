package atomic

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestAtomicNumber_SignedInt(t *testing.T) {
	num := NewAtomicNumber[int](10)

	// Get & Set
	assert.Equal(t, 10, num.Get())
	num.Set(100)
	assert.Equal(t, 100, num.Get())

	// Add & Sub
	num.Add(10)
	assert.Equal(t, 110, num.Get())
	num.Sub(20)
	assert.Equal(t, 90, num.Get())

	// Mul & Div
	num.Mul(2)
	assert.Equal(t, 180, num.Get())
	num.Div(3)
	assert.Equal(t, 60, num.Get())

	// Div by zero
	num.Div(0)
	assert.Equal(t, 60, num.Get())

	// Comparisons
	assert.True(t, num.Eq(60))
	assert.False(t, num.Eq(59))
	assert.True(t, num.Gt(50))
	assert.False(t, num.Gt(60))
	assert.True(t, num.Gte(60))
	assert.True(t, num.Lt(70))
	assert.False(t, num.Lt(60))
	assert.True(t, num.Lte(60))
}

func TestAtomicNumber_UnsignedInt_Underflow(t *testing.T) {
	num := NewAtomicNumber[uint](10)

	num.Sub(20) // 10 - 20 => 0
	assert.Equal(t, uint(0), num.Get())
}

func TestAtomicNumber_Float(t *testing.T) {
	num := NewAtomicNumber[float64](100.0)

	num.Add(50.5)
	assert.Equal(t, 150.5, num.Get())

	num.Sub(50.0)
	assert.Equal(t, 100.5, num.Get())

	num.Mul(2.0)
	assert.Equal(t, 201.0, num.Get())

	num.Div(3.0)
	assert.Equal(t, 67.0, num.Get())
}

func TestAtomicNumber_RaceCondition(t *testing.T) {
	num := NewAtomicNumber[int](0)
	var wg sync.WaitGroup
	numGoroutines := 1000

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			num.Add(1)
			num.Sub(1)
			num.Add(1)
		}()
	}

	wg.Wait()

	assert.Equal(t, 1000, num.Get())
}
