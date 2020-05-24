package rtree

import (
	"github.com/maja42/vmath"
)

var noBounds = vmath.Rectf{
	Min: vmath.Vec2f{
		vmath.Infinity,
		vmath.Infinity,
	},
	Max: vmath.Vec2f{
		vmath.NegInfinity,
		vmath.NegInfinity,
	},
}

// node is an R-Tree element that contains sub-elements.
type node struct {
	// Contains either children (leaf = false) or items (leaf=true), but never both.
	children []*node
	items    []Item

	height int
	leaf   bool
	bounds vmath.Rectf
}

func newNode() *node {
	return &node{
		height: 1,
		leaf:   true,
		bounds: noBounds,
	}
}

// sorting:
type nodesByMinX []*node
type nodesByMinY []*node

type itemsByMinX []Item
type itemsByMinY []Item

func (a nodesByMinX) Len() int           { return len(a) }
func (a nodesByMinX) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a nodesByMinX) Less(i, j int) bool { return a[i].bounds.Min[0] < a[j].bounds.Min[0] }

func (a nodesByMinY) Len() int           { return len(a) }
func (a nodesByMinY) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a nodesByMinY) Less(i, j int) bool { return a[i].bounds.Min[1] < a[j].bounds.Min[1] }

func (a itemsByMinX) Len() int           { return len(a) }
func (a itemsByMinX) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a itemsByMinX) Less(i, j int) bool { return a[i].Bounds().Min[0] < a[j].Bounds().Min[0] }

func (a itemsByMinY) Len() int           { return len(a) }
func (a itemsByMinY) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a itemsByMinY) Less(i, j int) bool { return a[i].Bounds().Min[1] < a[j].Bounds().Min[1] }
