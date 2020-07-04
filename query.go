package rtree

import (
	"math"
	"sort"

	"github.com/maja42/vmath"
	"github.com/maja42/vmath/math32"
)

const maxInt = math.MaxInt32

// All returns all stored items.
// Returns nil if the tree is empty.
func (r *RTree) All() []Item {
	var items []Item
	r.addAllItemsN(r.root, &items, maxInt)
	return items
}

// SearchPos returns all items at the given position.
func (r *RTree) SearchPos(pos vmath.Vec2f) []Item {
	return r.search(vmath.Rectf{pos, pos}, false, maxInt)
}

// SearchPos returns all items at the given position.
// Stops searching after 'maxResults' have found.
func (r *RTree) SearchPosN(pos vmath.Vec2f, maxResults int) []Item {
	return r.search(vmath.Rectf{pos, pos}, false, maxResults)
}

// Search returns all items within the area.
// If mustCover is true, items are only returned if they are fully within the search area.
// If false, items are returned if they intersect the search area.
func (r *RTree) Search(area vmath.Rectf, mustCover bool) []Item {
	return r.search(area, mustCover, maxInt)
}

// Search returns all items within the area.
// Stops searching after 'maxResults' have found.
// If mustCover is true, items are only returned if they are fully within the search area.
// If false, items are returned if they intersect the search area.
func (r *RTree) SearchN(area vmath.Rectf, mustCover bool, maxResults int) []Item {
	return r.search(area, mustCover, maxResults)
}

func (r *RTree) search(area vmath.Rectf, mustCover bool, maxResults int) []Item {
	area = area.Normalize()
	if !area.Intersects(r.root.bounds) {
		return nil
	}

	var items []Item

	nodesToSearch := make([]*node, 1)
	nodesToSearch[0] = r.root
	for len(nodesToSearch) > 0 {
		node := popNode(&nodesToSearch)

		for _, child := range node.children {
			if !area.Intersects(child.bounds) {
				continue
			}
			if area.ContainsRectf(child.bounds) {
				r.addAllItemsN(child, &items, maxResults)
				if len(items) >= maxResults {
					return items
				}
			} else {
				nodesToSearch = append(nodesToSearch, child)
			}
		}
		for _, item := range node.items {
			if (mustCover && area.ContainsRectf(item.Bounds())) ||
				(!mustCover && area.Intersects(item.Bounds())) {
				items = append(items, item)
				if len(items) >= maxResults {
					return items
				}
			}
		}
	}
	return items
}

// SearchFiltered returns all items within the area that are filtered.
// If 'filter' returns false, the item is discarded.
// If mustCover is true, items are only returned if they are fully within the search area.
// If false, items are returned if they intersect the search area.
func (r *RTree) SearchFiltered(area vmath.Rectf, mustCover bool, filter FilterFunc) []Item {
	area = area.Normalize()
	if !area.Intersects(r.root.bounds) {
		return nil
	}

	var items []Item

	nodesToSearch := make([]*node, 1)
	nodesToSearch[0] = r.root
	for len(nodesToSearch) > 0 {
		node := popNode(&nodesToSearch)

		for _, child := range node.children {
			if !area.Intersects(child.bounds) {
				continue
			}
			if area.ContainsRectf(child.bounds) {
				r.addAllFilteredItems(child, &items, filter)
			} else {
				nodesToSearch = append(nodesToSearch, child)
			}
		}
		for _, item := range node.items {
			if !filter(item) {
				continue
			}
			if (mustCover && area.ContainsRectf(item.Bounds())) ||
				(!mustCover && area.Intersects(item.Bounds())) {
				items = append(items, item)
			}
		}
	}
	return items
}

func (r *RTree) addAllItemsN(root *node, items *[]Item, maxLen int) {
	nodesToSearch := make([]*node, 1)
	nodesToSearch[0] = root
	for len(nodesToSearch) > 0 {
		node := popNode(&nodesToSearch)

		*items = append(*items, node.items...)
		if len(*items) >= maxLen {
			*items = (*items)[:maxLen]
			return
		}
		nodesToSearch = append(nodesToSearch, node.children...)
	}
}

func (r *RTree) addAllFilteredItems(root *node, items *[]Item, filter FilterFunc) {
	nodesToSearch := make([]*node, 1)
	nodesToSearch[0] = root
	for len(nodesToSearch) > 0 {
		node := popNode(&nodesToSearch)
		nodesToSearch = append(nodesToSearch, node.children...)

		for _, item := range node.items {
			if filter(item) {
				*items = append(*items, item)
			}
		}
	}
}

// Intersects returns true if there are any items overlapping with the given area.
// Touching rectangles where floats are exactly equal are not considered to intersect.
func (r *RTree) Intersects(area vmath.Rectf) bool {
	area = area.Normalize()
	if !area.Intersects(r.root.bounds) {
		return false
	}
	nodesToSearch := make([]*node, 1)
	nodesToSearch[0] = r.root
	for len(nodesToSearch) > 0 {
		node := popNode(&nodesToSearch)

		for _, child := range node.children {
			if !area.Intersects(child.bounds) {
				continue
			}
			if area.ContainsRectf(child.bounds) {
				return true
			} else {
				nodesToSearch = append(nodesToSearch, child)
			}
		}
		for _, item := range node.items {
			if area.Intersects(item.Bounds()) {
				return true
			}
		}
	}
	return false
}

// NearestNeighbor returns the item that is closest to the given position.
// Returns nil if the tree is empty.
func (r *RTree) NearestNeighbor(pos vmath.Vec2f) Item {
	item, _ := r.nearestNeighbor(pos, r.root, nil, math32.Infinity)
	return item
}

// NearestNeighbor returns the item that is closest to the given position but within the given max. distance.
// Returns nil if the tree is empty or if there are no items within the given distance.
func (r *RTree) NearestNeighborWithin(pos vmath.Vec2f, maxDistance float32) Item {
	maxSqDist := maxDistance * maxDistance
	item, _ := r.nearestNeighbor(pos, r.root, nil, maxSqDist)
	return item
}

func (r *RTree) nearestNeighbor(pos vmath.Vec2f, node *node, nearest Item, nearestSqDist float32) (Item, float32) {
	if node.leaf {
		for _, item := range node.items {
			itemDist := item.Bounds().SquarePointDistance(pos)
			if itemDist < nearestSqDist {
				nearestSqDist = itemDist
				nearest = item
			}
		}
		return nearest, nearestSqDist
	}

	// Sort nodes to visit most promising children first
	children := sortNodesByDistance(pos, node.children)
	// Prune nodes that can't contain the nearest neighbour
	children = pruneNodes(pos, nearestSqDist, children)

	for idx, child := range children.nodes {
		dist := children.sqDistances[idx]
		if dist > nearestSqDist {
			break
		}
		newNearest, newNearestSqDist := r.nearestNeighbor(pos, child, nearest, nearestSqDist)
		if newNearestSqDist < nearestSqDist {
			nearest = newNearest
			nearestSqDist = newNearestSqDist
		}
	}
	return nearest, nearestSqDist
}

// sortNodesByDistance reorders the node slice by distance to the given position.
// returns the sorted nodes and their squared distance.
func sortNodesByDistance(pos vmath.Vec2f, nodes []*node) nodesByDistance {
	// Note: sorting is done in-place --> copy nodes and leave original data-structure as-is.
	// This ensures that NearestNeighbor search can be handled as a read-only operation.

	sorted := nodesByDistance{
		nodes:       make([]*node, len(nodes), len(nodes)),
		sqDistances: make([]float32, len(nodes)),
	}
	copy(sorted.nodes, nodes)

	for i, nod := range sorted.nodes {
		sorted.sqDistances[i] = nod.bounds.SquarePointDistance(pos)
	}
	sort.Sort(sorted)
	return sorted
}

// pruneNodes removes all nodes that cannot contain contain the nearest neighbour.
func pruneNodes(pos vmath.Vec2f, nearestSqDist float32, sortedNodes nodesByDistance) nodesByDistance {
	// Calculate the min. distance within which it's guaranteed that there is an item
	minMinMaxDist := nearestSqDist
	for _, node := range sortedNodes.nodes {
		minMaxDist := minMaxDist(pos, node.bounds)
		minMinMaxDist = math32.Min(minMaxDist, minMinMaxDist)
	}
	// Remove all nodes that are farther away than 'minMinMaxDist'.
	cnt := 0
	for idx, node := range sortedNodes.nodes {
		dist := sortedNodes.sqDistances[idx]
		if dist <= minMinMaxDist {
			sortedNodes.nodes[cnt] = node
			sortedNodes.sqDistances[cnt] = dist
			cnt++
		}
	}
	return nodesByDistance{
		nodes:       sortedNodes.nodes[:cnt],
		sqDistances: sortedNodes.sqDistances[:cnt],
	}
}

// minMaxDist
// From all potential items within the given bounding box r,
// minMaxDist identifies the minimum distance within which at least one such item must exist.
// Returns the square of this minimum distance.
func minMaxDist(pos vmath.Vec2f, r vmath.Rectf) float32 {
	// Source: "Nearest Neighbor Queries" by N. Roussopoulos, S. Kelley and F. Vincent, ACM SIGMOD, pages 71-79, 1995.

	center := r.Min.Add(r.Max).MulScalar(0.5)

	rm := r.Max
	if pos[0] <= center[0] {
		rm[0] = r.Min[0]
	}
	if pos[1] <= center[1] {
		rm[1] = r.Min[1]
	}

	rM := r.Max
	if pos[0] >= center[0] {
		rM[0] = r.Min[0]
	}
	if pos[1] >= center[1] {
		rM[1] = r.Min[1]
	}

	// calculate distance S from pos to furthest corner of r
	d := pos.Sub(rM)
	S := d.Dot(d)

	dm := pos.Sub(rm)
	dM := pos.Sub(rM)
	d = vmath.Vec2f{
		S - dM[0]*dM[0] + dm[0]*dm[0],
		S - dM[1]*dM[1] + dm[1]*dm[1],
	}
	return math32.Min(d[0], d[1])
}

// IterateAllItems calls the provided function for every stored item until true (=abort) is returned.
// The order in which items are iterated is undefined.
func (r *RTree) IterateItems(fn func(item Item) bool) {
	nodesToSearch := make([]*node, 1)
	nodesToSearch[0] = r.root
	for len(nodesToSearch) > 0 {
		node := popNode(&nodesToSearch)

		for _, item := range node.items {
			if fn(item) {
				return
			}
		}
		nodesToSearch = append(nodesToSearch, node.children...)
	}
}

// IterateInternalNodes calls the provided function for every internal tree node until true (=abort) is returned.
// The order in which nodes are iterated is undefined.
// This function is useful for graphically visualizing the R-Tree internals.
func (r *RTree) IterateInternalNodes(fn func(bounds vmath.Rectf, height int, leaf bool) bool) {
	nodesToSearch := make([]*node, 1)
	nodesToSearch[0] = r.root
	for {
		if len(nodesToSearch) == 0 {
			return
		}
		node := popNode(&nodesToSearch)

		if fn(node.bounds, node.height, node.leaf) {
			return
		}
		nodesToSearch = append(nodesToSearch, node.children...)
	}
}

// Height returns the height of the R-Tree.
// This function is useful for graphically visualizing the R-Tree internals.
func (r *RTree) Height() int {
	if r.root == nil {
		return 0
	}
	return r.root.height
}

// Bounds returns the bounding box of all items.
// Returns an infinitely small bounding box if there are no items.
func (r *RTree) Bounds() vmath.Rectf {
	return r.root.bounds
}

// Size returns the total number of stored items.
func (r *RTree) Size() int {
	cnt := 0
	nodesToSearch := make([]*node, 1)
	nodesToSearch[0] = r.root
	for len(nodesToSearch) > 0 {
		node := popNode(&nodesToSearch)
		nodesToSearch = append(nodesToSearch, node.children...)
		cnt += len(node.items)
	}
	return cnt
}
