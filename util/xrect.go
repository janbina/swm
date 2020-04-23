package util

import (
	"github.com/BurntSushi/xgbutil/xrect"
	"math"
)

// MinMovement returns how much we need to move rect on x and y axis so it overlaps
// with some of rects by at least minOverlap pixels on both axis
// That movement will be minimal - to the closest rect from rects
func MinMovement(rect xrect.Rect, rects []xrect.Rect, minOverlap int) (x, y int) {
	minMovement, minX, minY := math.MaxInt32, 0, 0
	minOverlapX := min(rect.Width(), minOverlap)
	minOverlapY := min(rect.Height(), minOverlap)

	for _, test := range rects {
		minOverlapX := min(test.Width(), minOverlapX)
		minOverlapY := min(test.Height(), minOverlapY)

		x := neededMovement(rect.X(), rect.Width(), test.X(), test.Width(), minOverlapX)
		y := neededMovement(rect.Y(), rect.Height(), test.Y(), test.Height(), minOverlapY)

		if x == 0 && y == 0 {
			return 0, 0
		}
		if abs(x) + abs(y) < minMovement {
			minMovement = abs(x) + abs(y)
			minX, minY = x, y
		}
	}

	return minX, minY
}

func neededMovement(a1, aS, b1, bS, minOverlap int) int {
	a2 := a1 + aS
	b2 := b1 + bS
	if b2 - minOverlap < a1 {
		// a is too far right
		return b2 - minOverlap - a1
	} else if a2 < b1 + minOverlap {
		// a is too far left
		return b1 + minOverlap - a2
	}
	return 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
