package internal

import (
	"container/heap"
	"testing"
)

type testItem struct {
	name string
	val  int
}

func newTestHeap() *MinHeap[*testItem] {
	return NewMinHeap(func(a, b *testItem) bool {
		return a.val < b.val // smaller val = lower priority
	})
}

func TestMinHeap(t *testing.T) {
	t.Run("basic push pop", func(t *testing.T) {
		h := newTestHeap()

		items := []*testItem{
			{name: "a", val: 3},
			{name: "b", val: 1},
			{name: "c", val: 2},
		}

		for _, it := range items {
			heap.Push(h, it)
		}

		if h.Len() != 3 {
			t.Fatalf("expected len 3, got %d", h.Len())
		}

		// Pop should return items in ascending order
		expected := []int{1, 2, 3}
		for i, exp := range expected {
			got := heap.Pop(h).(*testItem)
			if got.val != exp {
				t.Errorf("pop %d: expected %d, got %d", i, exp, got.val)
			}
		}
	})

	t.Run("Top", func(t *testing.T) {
		h := newTestHeap()
		heap.Push(h, &testItem{"x", 10})
		heap.Push(h, &testItem{"y", 5})

		top, ok := h.Top()
		if !ok {
			t.Fatalf("expected Top() to return true")
		}
		if top.val != 5 {
			t.Errorf("expected top to be 5, got %d", top.val)
		}
	})

	t.Run("PushBounded", func(t *testing.T) {
		h := newTestHeap()
		k := 3

		// Insert 5 items, only top 3 (largest vals) should remain
		for i := 1; i <= 5; i++ {
			h.PushBounded(&testItem{name: string(rune('a' + i)), val: i}, k)
		}

		if h.Len() != k {
			t.Fatalf("expected heap length %d, got %d", k, h.Len())
		}

		// Pop all, should get 3, 4, 5 in ascending order
		expected := []int{3, 4, 5}
		for _, exp := range expected {
			item := heap.Pop(h).(*testItem)
			if item.val != exp {
				t.Errorf("expected val %d, got %d", exp, item.val)
			}
		}
	})

	t.Run("PushBounded replaces lowest", func(t *testing.T) {
		h := newTestHeap()
		k := 2

		h.PushBounded(&testItem{"a", 10}, k)
		h.PushBounded(&testItem{"b", 20}, k)
		h.PushBounded(&testItem{"c", 15}, k) // should replace 10
		h.PushBounded(&testItem{"a", 10}, k) // should NOT replace anything

		if h.Len() != 2 {
			t.Fatalf("expected len 2, got %d", h.Len())
		}

		// Extract remaining values to verify they’re top-2 (15, 20)
		got := []int{heap.Pop(h).(*testItem).val, heap.Pop(h).(*testItem).val}
		if got[0] != 15 || got[1] != 20 {
			t.Errorf("expected top-2 = [15,20], got %+v", got)
		}
	})

	t.Run("empty Top", func(t *testing.T) {
		h := newTestHeap()
		_, ok := h.Top()
		if ok {
			t.Fatal("expected Top() to return false on empty heap")
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when calling PushBounded on nil receiver")
			}
		}()
		var h *MinHeap[*testItem]
		h.PushBounded(&testItem{"x", 1}, 2)
	})
}
