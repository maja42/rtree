package rtree

import (
	"github.com/maja42/vmath"
)

// All returns all stored items.
// Returns nil if the tree is empty.
func (r *RTree) All() []Item {
	var items []Item
	r.addAllItems(r.root, &items)
	return items
}

// Search returns all items within the area.
// If mustCover is true, items are only returned if they are fully within the search area.
// If false, items are returned if they intersect the search area.
func (r *RTree) Search(area vmath.Rectf, mustCover bool) []Item {
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
				r.addAllItems(child, &items)
			} else {
				nodesToSearch = append(nodesToSearch, child)
			}
		}
		for _, item := range node.items {
			if (mustCover && area.ContainsRectf(item.Bounds())) ||
				(!mustCover && area.Intersects(item.Bounds())) {
				items = append(items, item)
			}
		}
	}
	return items
}

// Search returns all items within the area that are not filtered.
// If mustCover is true, items are only returned if they are fully within the search area.
// If false, items are returned if they intersect the search area.
func (r *RTree) FilteredSearch(area vmath.Rectf, mustCover bool, filter FilterFunc) []Item {
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

func (r *RTree) addAllItems(root *node, items *[]Item) {
	nodesToSearch := make([]*node, 1)
	nodesToSearch[0] = root
	for len(nodesToSearch) > 0 {
		node := popNode(&nodesToSearch)

		*items = append(*items, node.items...)
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
