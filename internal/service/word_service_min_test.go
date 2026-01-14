package service

import (
	"testing"
)

// Test min function
func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"a < b", 1, 2, 1},
		{"a > b", 5, 3, 3},
		{"a == b", 4, 4, 4},
		{"negative a < b", -1, 2, -1},
		{"negative a > b", -1, -5, -5},
		{"zero a", 0, 5, 0},
		{"zero b", 5, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
