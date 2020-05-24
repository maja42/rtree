package rtree

import (
	"math/rand"
	"testing"

	"github.com/maja42/vmath"
)

const testTreeSize = 10000

type testItem struct {
	data   []byte
	bounds vmath.Rectf
}

func (i *testItem) Bounds() vmath.Rectf {
	return i.bounds
}

func BenchmarkInsert(b *testing.B) {
	tree, _ := newPrePopulatedTree(testTreeSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Insert(randomItem())
	}
}

func BenchmarkSearch(b *testing.B) {
	tree, items := newPrePopulatedTree(testTreeSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item := items[rand.Intn(len(items))]
		_ = tree.Search(item.Bounds(), false)
	}
}

func BenchmarkFilteredSearch(b *testing.B) {
	tree, items := newPrePopulatedTree(testTreeSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item := items[rand.Intn(len(items))]
		_ = tree.FilteredSearch(item.Bounds(), false, func(item Item) bool {
			return true
		})
	}
}

func BenchmarkRemove(b *testing.B) {
	tree, items := newPrePopulatedTree(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Remove(items[i], nil)
	}
}

func newPrePopulatedTree(size int) (*RTree, []Item) {
	tree := New(0)
	items := make([]Item, size)
	for i := 0; i < size; i++ {
		items[i] = randomItem()
	}
	tree.BulkLoad(items)
	return tree, items
}

func randomItem() *testItem {
	return &testItem{
		data:   make([]byte, rand.Intn(2048)), // simulate big structs
		bounds: randomRect(),
	}
}

func randomRect() vmath.Rectf {
	dim := float32(100)
	return vmath.Rectf{
		Min: vmath.Vec2f{
			rand.Float32() * dim,
			rand.Float32() * dim,
		},
		Max: vmath.Vec2f{
			rand.Float32() * dim,
			rand.Float32() * dim,
		},
	}.Normalize()
}
