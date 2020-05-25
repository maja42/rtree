package rtree

import (
	"testing"

	"github.com/maja42/vmath"
	"github.com/stretchr/testify/assert"
)

func TestMinMaxDist(t *testing.T) {
	pos := vmath.Vec2f{0, 0}

	// Rectangle around pos
	r := vmath.Rectf{
		Min: vmath.Vec2f{-1, -1},
		Max: vmath.Vec2f{4, 6},
	}
	expected := vmath.Vec2f{4, 1}.SquareLength()
	mmd := minMaxDist(pos, r)
	assert.Equal(t, expected, mmd)

	// Rectangle right of pos
	r = vmath.Rectf{
		Min: vmath.Vec2f{10, -4},
		Max: vmath.Vec2f{14, 20},
	}
	expected = vmath.Vec2f{14, -4}.SquareLength()
	mmd = minMaxDist(pos, r)
	assert.Equal(t, expected, mmd)

	// Rectangle left of pos
	r = vmath.Rectf{
		Min: vmath.Vec2f{-15, 0},
		Max: vmath.Vec2f{-10, 8},
	}
	expected = vmath.Vec2f{-10, 8}.SquareLength()
	mmd = minMaxDist(pos, r)
	assert.Equal(t, expected, mmd)

	// Rectangle below + left of pos
	r = vmath.Rectf{
		Min: vmath.Vec2f{-13, -16},
		Max: vmath.Vec2f{-3, -9},
	}
	expected = vmath.Vec2f{-13, -9}.SquareLength()
	mmd = minMaxDist(pos, r)
	assert.Equal(t, expected, mmd)
}
