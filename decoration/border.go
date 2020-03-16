package decoration

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
)

type Position int

const (
	Left Position = iota
	Right
	Top
	Bottom
)

type Border struct {
	position Position
	size int
	win *xwindow.Window
}

func CreateBorder(parent *xwindow.Window, position Position, size int, color uint32) (Decoration, error) {
	X := parent.X

	win, _ := xwindow.Create(X, parent.Id)
	win.Change(xproto.CwBackPixel, color)

	return Decoration(&Border{
		position: position,
		size:     size,
		win:      win,
	}), nil
}

func CreateBorders(parent *xwindow.Window, size int, color uint32) (Decorations, error) {
	dec := make(Decorations, 4)
	for i, p := range []Position{Left, Right, Top, Bottom} {
		b, err := CreateBorder(parent, p, size, color)
		if err != nil {
			return nil, err
		}
		dec[i] = b
	}
	return dec, nil
}

func (b *Border) ApplyRect(rect xrect.Rect) xrect.Rect {
	var x, y, w, h int

	switch b.position {
	case Left:
		x = rect.X()
		y = rect.Y()
		w = b.size
		h = rect.Height()
	case Top:
		x = rect.X()
		y = rect.Y()
		w = rect.Width()
		h = b.size
	case Right:
		x = rect.X() + rect.Width() - b.size
		y = rect.Y()
		w = b.size
		h = rect.Height()
	case Bottom:
		x = rect.X()
		y = rect.Y() + rect.Height() - b.size
		w = rect.Width()
		h = b.size
	}

	b.win.MoveResize(x, y, w, h)

	newRect := xrect.New(rect.X(), rect.Y(), rect.Width(), rect.Height())
	switch b.position {
	case Left:
		newRect.XSet(newRect.X() + b.size)
		newRect.WidthSet(newRect.Width() - b.size)
	case Top:
		newRect.YSet(newRect.Y() + b.size)
		newRect.HeightSet(newRect.Height() - b.size)
	case Right:
		newRect.WidthSet(newRect.Width() - b.size)
	case Bottom:
		newRect.HeightSet(newRect.Height() - b.size)
	}

	return newRect
}

func (b *Border) WidthNeeded() int {
	switch b.position {
	case Left, Right:
		return b.size
	default:
		return 0
	}
}

func (b *Border) HeightNeeded() int {
	switch b.position {
	case Top, Bottom:
		return b.size
	default:
		return 0
	}
}

func (b *Border) Map() {
	b.win.Map()
}

func (b *Border) Unmap() {
	b.win.Unmap()
}

func (b *Border) Destroy() {
	b.win.Destroy()
}
