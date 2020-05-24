package rtree

import (
	"math"
	"sort"
	"sync"

	"github.com/maja42/vmath"
)

type Item interface {
	//	Bounds() vmath.Rectf
}

type BoundsFunc func(item interface{}) vmath.Rectf

type EqualsFunc func(a, b Item) bool

type RTree struct {
	maxEntries, minEntries int // #entries within a single node
	bboxFn                 BoundsFunc
	root                   *node
}

// New creates a new, empty RTree for storing and querying 2D rectangles and points.
func New(boundsFunc BoundsFunc, maxEntries int) *RTree {
	if maxEntries <= 0 {
		// 16 entries for best performance and lowest memory overhead
		maxEntries = 16
	}
	maxEntries = vmath.Maxi(4, maxEntries)

	// min node fill is 40% for best performance
	r := &RTree{
		bboxFn:     boundsFunc,
		maxEntries: maxEntries,
		minEntries: vmath.Maxi(2, int(vmath.Ceil(float32(maxEntries)*0.4))),
	}
	r.Clear()
	return r
}

// Clear removes all items.
func (r *RTree) Clear() *RTree {
	r.root = newNode()
	return r
}

// Insert adds a single item.
func (r *RTree) Insert(item Item) *RTree {
	bbox := r.bboxFn(item)
	level := r.root.height - 1

	// determine best leaf node for new item and the path to get there
	leafNode, insertPath := r.chooseSubtree(bbox, r.root, level)
	leafNode.items = append(leafNode.items, item)
	extend(&leafNode.bounds, bbox)

	r.splitNodes(insertPath, level)

	// adjust bounding boxes along the insertion path
	r.adjustParentBBoxes(insertPath, bbox, level)
	return r
}

// BulkLoad inserts big root sets at once.
//
// Bulk insertion can be ~5-6 times faster than inserting items one by one.
// Subsequent search queries are also ~2-3 times faster.
//
// Note that when you do bulk insertion into an existing tree,
// it bulk-loads the given root into a separate tree and inserts the smaller tree into the larger tree.
// This means that bulk insertion works very well for clustered root (where items in one update are close to each other),
// but makes query performance worse if the root is scattered.
func (r *RTree) BulkLoad(items []Item) *RTree {
	if len(items) < r.minEntries {
		for _, item := range items {
			r.Insert(item)
		}
		return r
	}

	newTree := r.build(items, 0, len(items)-1, 0)

	if len(r.root.children)+len(r.root.items) == 0 {
		r.root = newTree
	} else if r.root.height == newTree.height {
		r.splitRoot(r.root, newTree)
	} else {
		// insert the small tree into the large tree at appropriate level
		if r.root.height < newTree.height { // swap trees
			r.root, newTree = newTree, r.root
		}
		r.insertNode(newTree, r.root.height-newTree.height-1)
	}
	return r
}

// Remove the given item from the tree.
// equalsFn is optional. It is useful if you only have a copy of the originally inserted item.
func (r *RTree) Remove(item Item, equalsFn EqualsFunc) *RTree {
	bbox := r.bboxFn(item)

	var path []*node       // path to current node from top->bottom
	var childIndexes []int // last processed childIdx for each node on the path
	var parent *node
	var childIdx int

	goingUp := false

	// depth-first iterative tree traversal
	nod := r.root
	for nod != nil || len(path) > 0 {
		if nod == nil { // go up
			nod = popNode(&path)
			parent = r.root //
			if len(path) > 1 {
				parent = path[len(path)-1]
			}
			childIdx = popInt(&childIndexes) // continue at previous child
			goingUp = true
		}

		if nod.leaf { // check current node
			if removeChildItem(nod, item, equalsFn) { // item found
				r.condense(append(path, nod)) // remove empty nodes and update bounding boxes
				return r
			}
		}

		contained := nod.bounds.ContainsRectf(bbox)
		if !goingUp && !nod.leaf && contained { // go down
			// remember current position on this level:
			path = append(path, nod)
			childIndexes = append(childIndexes, childIdx)
			// continue at first child:
			childIdx = 0
			parent = nod
			nod = nod.children[0]
		} else if parent != nil { // go right
			nod = nil
			childIdx++
			if childIdx < len(parent.children) {
				nod = parent.children[childIdx]
			}
			goingUp = false
		} else { // nothing found
			nod = nil
		}
	}
	return r
}

// insertNode inserts the new node (and it's subtree) at the given level
func (r *RTree) insertNode(node *node, level int) {
	bbox := node.bounds

	// determine best node for new child and the path to get there
	leafNode, insertPath := r.chooseSubtree(bbox, r.root, level)
	leafNode.children = append(leafNode.children, node)
	extend(&leafNode.bounds, bbox)

	r.splitNodes(insertPath, level)

	// adjust bounding boxes along the insertion path
	r.adjustParentBBoxes(insertPath, bbox, level)
}

// splitNodes splits all overflowing nodes along the insertion path
func (r *RTree) splitNodes(insertPath []*node, level int) {
	for level >= 0 {
		entries := len(insertPath[level].children) + len(insertPath[level].items)
		if entries <= r.maxEntries {
			break
		}
		r.split(insertPath, level)
		level--
	}
}

// build recursively creates a new tree with the given items using an OMT (overlap minimizing top-down bulk loading) algorithm.
func (r *RTree) build(items []Item, left, right, height int) *node {
	count := float64(right - left + 1)
	max := float64(r.maxEntries)

	if count <= max { // create leaf
		node := newNode()
		node.items = append(node.items, items[left:right+1]...)
		calcBBox(node, r.bboxFn)
		return node
	}

	if height == 0 {
		height = int(math.Ceil(logN(count, max)))  //target height of resulting tree =  LOGmax(count)
		maxCap := math.Pow(max, float64(height-1)) // total capacity in the resulting tree
		max = math.Ceil(count / maxCap)            // target number of root entries to maximize storage utilization
	}

	node := newNode()
	node.leaf = false
	node.height = height

	// split the items into 'max' groups, where each group is mostly square
	// This is done by grouping all nodes by their x-coordinate into 'grpX' groups.
	// The resulting groups are then each grouped again by their y-coordinate into 'grpY' groups.

	grpY := int(math.Ceil(count / max))
	grpX := grpY * int(math.Ceil(math.Sqrt(max)))

	groupItems(items, left, right, grpX, true, r.bboxFn)

	var wg sync.WaitGroup
	var m sync.Mutex

	for i := left; i <= right; i += grpX {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			right2 := vmath.Mini(i+grpX-1, right)
			// sort group [i, right2] again, but now by y
			groupItems(items, i, right2, grpY, false, r.bboxFn)

			for j := i; j <= right2; j += grpY {
				right3 := vmath.Mini(j+grpY-1, right2)
				// group [j, right3] is now nearly square; add it recursively
				sub := r.build(items, j, right3, height-1)
				m.Lock()
				node.children = append(node.children, sub)
				m.Unlock()
			}
		}(i)
	}
	wg.Wait()
	calcBBox(node, r.bboxFn)
	return node
}

// chooseSubtree finds the node that is best suited for the new entry.
// Returns the node and the path to find it. The found node is not part of the path.
// level defines the height at which the node should be inserted (in case of bulk-loads, where whole sub-trees are inserted).
func (r *RTree) chooseSubtree(bbox vmath.Rectf, root *node, level int) (*node, []*node) {
	path := make([]*node, 0)

	subNode := root
	for {
		path = append(path, subNode)

		if subNode.leaf || len(path)-1 == level {
			// leaf node found or maximum search-depth reached
			break
		}

		minArea := vmath.Infinity
		minEnlargement := vmath.Infinity
		var nextSubNode *node

		for _, child := range subNode.children {
			area := child.bounds.Area()
			enlargement := enlargedArea(bbox, child.bounds) - area

			// choose entry with the least area enlargement
			if enlargement < minEnlargement {
				minEnlargement = enlargement
				minArea = vmath.Min(minArea, area)
				nextSubNode = child
				continue
			}
			// otherwise choose one with the smallest area
			if enlargement == minEnlargement {
				if area < minArea {
					minArea = area
					nextSubNode = child
				}
			}
		}
		subNode = nextSubNode
	}
	return subNode, path
}

// split overflowed node at index 'level' into two
func (r *RTree) split(insertPath []*node, level int) {
	node := insertPath[level]
	min := r.minEntries
	max := len(node.children) + len(node.items)

	r.chooseSplitAxis(node, min, max)
	splitIndex := r.chooseSplitIndex(node, min, max)

	newNode := newNode()
	newNode.height = node.height
	newNode.leaf = node.leaf

	if node.leaf {
		newNode.items = append(newNode.items, node.items[splitIndex:]...)
		node.items = node.items[:splitIndex]
	} else {
		newNode.children = append(newNode.children, node.children[splitIndex:]...)
		node.children = node.children[:splitIndex]
	}

	calcBBox(node, r.bboxFn)
	calcBBox(newNode, r.bboxFn)

	if level > 0 {
		insertPath[level-1].children = append(insertPath[level-1].children, newNode)
	} else {
		r.splitRoot(node, newNode)
	}
}

// splitRoot splits the current root node into two.
func (r *RTree) splitRoot(a, b *node) {
	newHeight := r.root.height + 1
	r.root = newNode()
	r.root.children = []*node{a, b}

	r.root.height = newHeight
	r.root.leaf = false
	calcBBox(r.root, r.bboxFn)
}

// chooseSplitIndex finds the index at which the nodes' children should be split.
// The node's children are already sorted by the ideal split axis.
// min is the minimum number of entries in a node. count is the current number entries.
func (r *RTree) chooseSplitIndex(node *node, min, count int) int {
	minOverlap := vmath.Infinity
	minArea := vmath.Infinity

	idx := count - min // default index = maximum
	for i := min; i <= count-min; i++ {
		bbox1 := calcSubBBox(node, 0, i, r.bboxFn)
		bbox2 := calcSubBBox(node, i, count, r.bboxFn)

		overlap := intersectionArea(bbox1, bbox2)
		area := bbox1.Area() + bbox2.Area()

		if overlap < minOverlap {
			// choose distribution with minimum overlap
			minOverlap = overlap
			minArea = vmath.Min(area, minArea)
			idx = i
		} else if overlap == minOverlap {
			// otherwise choose distribution with minimum area
			if area < minArea {
				minArea = area
				idx = i
			}
		}
	}
	return idx
}

// chooseSplitAxis sorts node entries by the best axis for the split
func (r *RTree) chooseSplitAxis(nod *node, min, max int) {
	// determine sorting algorithm for each axis:
	var sortMinX, sortMinY sort.Interface
	if nod.leaf {
		sortMinX = itemsByMinX{nod.items, r.bboxFn}
		sortMinY = itemsByMinY{nod.items, r.bboxFn}
	} else {
		sortMinX = nodesByMinX(nod.children)
		sortMinY = nodesByMinY(nod.children)
	}

	sort.Sort(sortMinX)
	xMargin := r.allDistMargin(nod, min, max)
	sort.Sort(sortMinY)
	yMargin := r.allDistMargin(nod, min, max)

	// if total distributions margin value is minimal for x, sort by minX
	// otherwise it's already sorted by minY
	if xMargin < yMargin {
		sort.Sort(sortMinX)
	}
}

// allDistMargin calculates the total margin of all possible split distributions, where each node is at least min full
// The result can be used as a heuristic to determine how to split nodes.
func (r *RTree) allDistMargin(nod *node, min, max int) float32 {
	leftBBox := calcSubBBox(nod, 0, min, r.bboxFn)
	rightBBox := calcSubBBox(nod, max-min, max, r.bboxFn)

	margin := bboxMargin(leftBBox) + bboxMargin(rightBBox)

	for i := min; i < max-min; i++ {
		if nod.leaf {
			child := nod.items[i]
			extend(&leftBBox, r.bboxFn(child))
		} else {
			child := nod.children[i]
			extend(&leftBBox, child.bounds)
		}
		margin += bboxMargin(leftBBox)
	}

	for i := max - min - 1; i >= min; i-- {
		if nod.leaf {
			child := nod.items[i]
			extend(&rightBBox, r.bboxFn(child))
		} else {
			child := nod.children[i]
			extend(&rightBBox, child.bounds)
		}
		margin += bboxMargin(rightBBox)
	}
	return margin
}

// adjustParentBBoxes adds the new bbox to all bounding boxes along the insertion path
func (r *RTree) adjustParentBBoxes(insertPath []*node, bbox vmath.Rectf, level int) {
	for i := level; i >= 0; i-- {
		extend(&insertPath[i].bounds, bbox)
	}
}

// condense removes all empty nodes from the given path and updates the bounding boxes.
func (r *RTree) condense(path []*node) {
	for i := len(path) - 1; i >= 0; i-- {
		item := path[i]
		itemCount := len(item.children) + len(item.items)
		if itemCount == 0 { // empty
			if i > 0 {
				parent := path[i-1]
				removeChildNode(parent, item)
			} else { // tree is empty
				r.Clear()
			}
		} else {
			calcBBox(item, r.bboxFn)
		}
	}
}

// removeChildItem removes a child item from its direct parent.
// Returns true if the child was found and removed.
func removeChildItem(parent *node, child Item, equalsFn EqualsFunc) bool {
	for idx, item := range parent.items {
		var found bool
		if equalsFn == nil {
			found = child == item
		} else {
			found = equalsFn(child, item)
		}
		if found {
			parent.items = append(parent.items[:idx], parent.items[idx+1:]...) //remove item
			return true
		}
	}
	return false
}

// removeChildNode removes a child node from its direct parent.
func removeChildNode(parent, child *node) {
	for idx, node := range parent.children {
		if node == child {
			parent.children = append(parent.children[:idx], parent.children[idx+1:]...)
			return
		}
	}
}

// groupItems partially sorts the item slice into groups of n unsorted items.
// The groups are sorted between each other.
// If xDim is true, the MinX position is used for sorting, otherwise MinY is used.
// Combines quickselect with a non-recursive divide & conquer algorithm.
func groupItems(items []Item, leftIdx, rightIdx, groupSize int, xDim bool, bboxFn BoundsFunc) {
	stack := []int{leftIdx, rightIdx}
	for len(stack) > 0 {
		rightIdx, leftIdx = popInt(&stack), popInt(&stack)

		size := rightIdx - leftIdx
		if size <= groupSize {
			continue
		}

		groups := float64(size) / float64(groupSize)
		pivot := int(math.Ceil(groups/2)) * groupSize // center group
		if xDim {
			//quickselectFloyd(itemsByMinX{items[leftIdx:rightIdx+1], bboxFn}, pivot)
			quickselect(itemsByMinX{items[leftIdx : rightIdx+1], bboxFn}, pivot)
			//nth.Element(itemsByMinX{items[leftIdx:rightIdx+1], bboxFn}, pivot)
		} else {
			//quickselectFloyd(itemsByMinY{items[leftIdx:rightIdx+1], bboxFn}, pivot)
			quickselect(itemsByMinY{items[leftIdx : rightIdx+1], bboxFn}, pivot)
			//nth.Element(itemsByMinY{items[leftIdx:rightIdx+1], bboxFn}, pivot)
		}
		pivot += leftIdx
		// repeat on the left and right side of the pivot point
		stack = append(stack, leftIdx, pivot, pivot, rightIdx)
	}
}

// popNode removes and returns the last slice entry.
func popNode(nodes *[]*node) *node {
	length := len(*nodes)
	node := (*nodes)[length-1]
	*nodes = (*nodes)[:length-1]
	return node
}

// popInt removes and returns the last slice entry.
func popInt(ints *[]int) int {
	length := len(*ints)
	i := (*ints)[length-1]
	*ints = (*ints)[:length-1]
	return i
}

// calculate node's bbox from bboxes of its children
func calcBBox(node *node, bboxFn BoundsFunc) {
	node.bounds = calcSubBBox(node, 0, len(node.children)+len(node.items), bboxFn)
}

// calcSubBBox calculates the bbox for all entries in slice [start:end].
func calcSubBBox(node *node, start, end int, bboxFn BoundsFunc) vmath.Rectf {
	bbox := noBounds
	if node.leaf {
		for _, item := range node.items[start:end] {
			extend(&bbox, bboxFn(item))
		}
	} else {
		for _, child := range node.children[start:end] {
			extend(&bbox, child.bounds)
		}
	}
	return bbox
}

// intersectionArea returns the area after merging the two given boxes.
func intersectionArea(a, b vmath.Rectf) float32 {
	return a.Merge(b).Area()
}

// enlargedArea calculates the new area of a bounding box when adding a child.
func enlargedArea(bbox, newChild vmath.Rectf) float32 {
	width := vmath.Max(newChild.Max[0], bbox.Max[0]) - vmath.Min(newChild.Min[0], bbox.Min[0])
	height := vmath.Max(newChild.Max[1], bbox.Max[1]) - vmath.Min(newChild.Min[1], bbox.Min[1])
	return width * height
}

func extend(a *vmath.Rectf, b vmath.Rectf) {
	*a = a.Merge(b)
}

// bboxMargin returns the bbox's sum of width and height.
func bboxMargin(bbox vmath.Rectf) float32 {
	return (bbox.Max[0] - bbox.Min[0]) + (bbox.Max[1] - bbox.Min[1])
}

func logN(v, base float64) float64 {
	return math.Log(v) / math.Log(base)
}
