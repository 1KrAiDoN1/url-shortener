package random

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRandomString(t *testing.T) {
	test := make(map[int]struct {
		name string
		size int
	}, 100)
	for i := 1; i <= 100; i++ {
		test[i] = struct {
			name string
			size int
		}{
			name: fmt.Sprintf("test_%d", i),
			size: i,
		}
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			str1 := NewRandomString(tt.size)
			str2 := NewRandomString(tt.size)

			assert.Len(t, str1, tt.size)
			assert.Len(t, str2, tt.size)

			// Check that two generated strings are different
			// This is not an absolute guarantee that the function works correctly,
			// but this is a good heuristic for a simple random generator.
			assert.NotEqual(t, str1, str2)
		})
	}
}
