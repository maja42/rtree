package rtree

import (
	"math/rand"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQuickSelect(t *testing.T) {
	arr := []int{65, 28, 59, 52, 21, 56, 22, 95, 50, 12, 90, 53, 28, 54, 39}
	pivot := 8
	quickselect(sort.IntSlice(arr), pivot)
	assertQuickSelectResult(t, arr, pivot)
}

func TestQuickSelect_BruteForce(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	testCases := 500

	for tc := 0; tc < testCases; tc++ {
		t.Run("test case "+strconv.Itoa(tc), func(t *testing.T) {
			testSize := 1 + rand.Intn(2048)
			arr := make([]int, testSize)
			for i := 0; i < testSize; i++ {
				arr[i] = rand.Int()
			}

			pivot := rand.Intn(testSize)
			quickselect(sort.IntSlice(arr), pivot)

			if !assertQuickSelectResult(t, arr, pivot) {
				t.Logf("Pivot: %d (=%d), Data: %v", pivot, arr[pivot], arr)
			}
		})
	}
}

func assertQuickSelectResult(t *testing.T, arr []int, pivot int) bool {
	t.Helper()

	pivotVal := arr[pivot]
	for i := 0; i < pivot; i++ {
		if !assert.LessOrEqualf(t, arr[i], pivotVal, "Index %d (=%d) > pivot", i, arr[i]) {
			return false
		}
	}
	for i := pivot + 1; i < len(arr)-1; i++ {
		if !assert.GreaterOrEqualf(t, arr[i], pivotVal, "Index %d (=%d) < pivot", i, arr[i]) {
			return false
		}
	}
	return true
}

func BenchmarkQuickSelect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		arr := makeTestData()
		b.StartTimer()
		quickselect(arr, arr.Len()/2)
	}
}

// github.com/wangjohn/quickselect
//func BenchmarkQuickSelect_quickselect(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		b.StopTimer()
//		arr := makeTestData()
//		b.StartTimer()
//		_ = quicksel.QuickSelect(arr, arr.Len()/2)
//	}
//}

// github.com/keegancsmith/nth
//func BenchmarkQuickSelect_nth(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		b.StopTimer()
//		arr := makeTestData()
//		b.StartTimer()
//		nth.Element(arr, arr.Len()/2)
//	}
//}

type bigTestItem struct {
	Data []byte
	Num  int
}

func makeTestData() sort.Interface {
	arr := make([]bigTestItem, 5000)
	for i := 0; i < len(arr); i++ {
		arr[i] = randBigType()
	}
	return bigTestItemSlice(arr)
}

func randBigType() bigTestItem {
	return bigTestItem{
		Data: make([]byte, rand.Intn(2048)), // simulate big structs
		Num:  rand.Int(),
	}
}

type bigTestItemSlice []bigTestItem

func (t bigTestItemSlice) Len() int {
	return len(t)
}

func (t bigTestItemSlice) Less(i, j int) bool {
	return t[i].Num < t[j].Num
}

func (t bigTestItemSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
