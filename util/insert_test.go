package util

import (
	"testing"
)

func TestInsert(t *testing.T) {
	data := []int64{23, 89, 100}
	counter := []int64{0, 0, 0}
	InsertLeft(data, counter, 1)
	InsertLeft(data, counter, 24)
	InsertLeft(data, counter, 100)
	InsertLeft(data, counter, 101)
	InsertLeft(data, counter, 102)
	InsertLeft(data, counter, 100)
	InsertLeft(data, counter, 23)
	// Expect: 2, 1, 4
	if counter[0] != 2 || counter[1] != 1 || counter[2] != 4 {
		t.Error(counter)
	}

}
