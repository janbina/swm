package decoration

import "github.com/BurntSushi/xgbutil/xrect"

type Decoration interface {
	ApplyRect(rect xrect.Rect) xrect.Rect
	WidthNeeded() int
	HeightNeeded() int
	Map()
	Unmap()
	Destroy()
}

type Decorations []Decoration

func (d *Decorations) ApplyRect(rect xrect.Rect) xrect.Rect {
	for _, decoration := range *d {
		rect = decoration.ApplyRect(rect)
	}
	return rect
}

func (d *Decorations) WidthNeeded() int {
	sum := 0
	for _, decoration := range *d {
		sum += decoration.WidthNeeded()
	}
	return sum
}

func (d *Decorations) HeightNeeded() int {
	sum := 0
	for _, decoration := range *d {
		sum += decoration.HeightNeeded()
	}
	return sum
}

func (d *Decorations) Map() {
	for _, decoration := range *d {
		decoration.Map()
	}
}

func (d *Decorations) Unmap() {
	for _, decoration := range *d {
		decoration.Unmap()
	}
}

func (d *Decorations) Destroy() {
	for _, decoration := range *d {
		decoration.Destroy()
	}
}
