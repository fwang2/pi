package util

import (
	"testing"
)

func TestSortedQueue(t *testing.T) {
	data := []Item{
		{"Alice", 23},
		{"Eve", 2},
		{"Bob", 25},
		{"Qiqi", 35},
		{"Yang", 96},
	}

	q := NewSortedQueue(3)

	for _, d := range data {
		q.Put(d)
	}
	if len(q.items) != 3 {
		t.Errorf("Incorrect length: %d\n", len(q.items))
	}

	if q.items[0].Val != 25 || q.items[1].Val != 35 || q.items[2].Val != 96 {
		t.Errorf("Incorrect content %v\n", q.items)
	}

	q = NewSortedQueue(1)

	for _, d := range data {
		q.Put(d)
	}
	if len(q.items) != 1 {
		t.Errorf("Incorrect length: %d\n", len(q.items))
	}

	if q.items[0].Val != 96 {
		t.Errorf("Incorrect content %v\n", q.items)
	}
}
