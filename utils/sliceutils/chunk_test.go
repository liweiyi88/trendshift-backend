package sliceutils

import (
	"testing"
)

func TestChunk(t *testing.T) {
	number := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	chunks := Chunk[int](number, 3)

	if len(chunks) != 4 {
		t.Errorf("expect %d chunks but got: %d", 4, len(chunks))
	}
}
