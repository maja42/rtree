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

////quickselect partitions the slice elements such that:
////     all elements with idx < pivot have a value smaller than the pivot element and
////     all elements with idx > pivot have a bigger value than the pivot element
////It uses the Floyd-Rivest selection algorithm for optimal performance.
//func quickselectFloyd(arr sort.Interface, pivot int) {
//	quickselectStep(arr, 0, arr.Len()-1, pivot)
//}
//
//func quickselectStep(srt sort.Interface, left, right, pivot int) {
//	var fk = float64(pivot)
//
//	for right > left {
//		if right-left > 600 {
//			// Use select recursively to sample a smaller set of size s.
//			// The arbitrary constants 600 and 0.5 are used in the original version to minimize execution time.
//			var n = float64(right - left + 1)
//			var i = float64(pivot - left + 1)
//			var z = math.Log(n)
//			var s = 0.5 * math.Exp(2*z/3)
//			var sd = 0.5 * math.Sqrt(z*s*(n-s)/n) * sign(i-n/2)
//			var newLeft = vmath.Maxi(left, int(math.Floor(fk-i*s/n+sd)))
//			var newRight = vmath.Mini(right, int(math.Floor(fk+(n-i)*s/n+sd)))
//			quickselectStep(srt, newLeft, newRight, pivot)
//		}
//
//		// partition the elements between left and right around t
//		var pvtIdx = pivot
//		var i = left
//		var j = right
//
//		srt.Swap(left, pivot)
//		pvtIdx = left
//		if srt.Less(pvtIdx, right) {
//			srt.Swap(left, right)
//			pvtIdx = right
//		}
//
//		for i < j {
//			srt.Swap(i, j)
//			if i == pvtIdx {
//				pvtIdx = j
//			} else if j == pvtIdx {
//				pvtIdx = i
//			}
//			i++
//			j--
//			for srt.Less(i, pvtIdx) {
//				i++
//			}
//			for srt.Less(pvtIdx, j) {
//				j--
//			}
//		}
//
//		if srt.Less(left, pvtIdx) == srt.Less(pvtIdx, left) { // left == pivot
//			srt.Swap(left, j)
//		} else {
//			j++
//			srt.Swap(j, right)
//		}
//
//		// Adjust left and right towards the boundaries of the subset
//		// containing the (k − left + 1)th smallest element.
//		if j <= pivot {
//			left = j + 1
//		}
//		if pivot <= j {
//			right = j - 1
//		}
//	}
//}
//
//func sign(f float64) float64 {
//	if math.Signbit(f) {
//		return 1
//	}
//	return -1
//}
