package rtree

import (
	"math/rand"
	"sort"
)

// quickselect performs a partial sort, ensuring that all elements before 'n' have a smaller value,
// and all elements after 'n' are bigger. This is equivalent to finding the nth smallest element.
//
// The used algorithm is a naive approach, but turned out to have better performance than existing libraries
// like "github.com/keegancsmith/nth" and "github.com/wangjohn/quickselect".
// It also provided better performance than a custom implementation using the Floyd-Rivest selection algorithm
// which is explained here: https://en.wikipedia.org/wiki/Floyd%E2%80%93Rivest_algorithm
func quickselect(a sort.Interface, n int) {
	first := 0
	last := a.Len() - 1
	for {
		guess := rand.Intn(last-first+1) + first
		pivotIndex := partition(a, first, last, guess)
		if n == pivotIndex { // found nth element
			return
		} else if n < pivotIndex { // nth element is on the left side
			last = pivotIndex - 1
		} else { // nth element is on the right side
			first = pivotIndex + 1
		}
	}
}

// partition moves all elements smaller than the pivot to its left, and all bigger values to its right.
// Returns the new position of the pivot.
func partition(a sort.Interface, firstIdx, lastIdx, pivotIdx int) int {
	a.Swap(firstIdx, pivotIdx) // move to front
	pivotIdx = firstIdx

	left, right := firstIdx+1, lastIdx

	for left <= right { // move to center
		for left <= lastIdx && a.Less(left, pivotIdx) {
			left++
		}
		for right >= pivotIdx && a.Less(pivotIdx, right) {
			right--
		}
		if left <= right {
			a.Swap(left, right)
			left++
			right--
		}
	}
	a.Swap(pivotIdx, right) // swap into right place
	return right
}
