package generic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	a := assert.New(t)
	s0 := NewSet[int]()
	s1 := NewSet(1, 2)
	s2 := NewSet(2, 3)
	a.Equal(NewSet(1), s1.Subtract(s2))
	a.Equal(NewSet(1, 2, 3), s1.Union(s2))
	a.Equal(NewSet(2), s1.Intersect(s2))
	a.True(s1.Contains(1))
	a.False(s1.Contains(3))
	a.False(s0.Contains(1))
}
