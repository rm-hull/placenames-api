package internal

import "container/heap"

type MinHeap[T any] struct {
	data []T
	less func(a, b T) bool // defines heap ordering: true if a < b
}

func NewMinHeap[T any](less func(a, b T) bool) *MinHeap[T] {
	return &MinHeap[T]{less: less}
}

func (h MinHeap[T]) Len() int {
	return len(h.data)
}

func (h MinHeap[T]) Less(i, j int) bool {
	return h.less(h.data[i], h.data[j])
}

func (h MinHeap[T]) Swap(i, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
}

func (h *MinHeap[T]) Push(x any) {
	h.data = append(h.data, x.(T))
}

func (h *MinHeap[T]) Pop() any {
	old := h.data
	n := len(old)
	x := old[n-1]
	h.data = old[:n-1]
	return x
}

func (h *MinHeap[T]) Top() (T, bool) {
	var zero T
	if len(h.data) == 0 {
		return zero, false
	}
	return h.data[0], true
}

func (h *MinHeap[T]) Items() []T {
	return h.data
}

func (h *MinHeap[T]) PushBounded(x T, k int) {
	if len(h.data) < k {
		heap.Push(h, x)
	} else if h.less(h.data[0], x) {
		heap.Pop(h)
		heap.Push(h, x)
	}
}
