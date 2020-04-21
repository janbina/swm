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
	position       Position
	size           int
	win            *xwindow.Window
	colorNormal    uint32
	colorActive    uint32
	colorAttention uint32
}

func CreateBorder(
	parent *xwindow.Window, position Position, size int, colorNormal, colorActive, colorAttention uint32,
) Decoration {
	X := parent.X

	win, _ := xwindow.Create(X, parent.Id)
	win.Change(xproto.CwBackPixel, colorNormal)

	return &Border{
		position:       position,
		size:           size,
		win:            win,
		colorNormal:    colorNormal,
		colorActive:    colorActive,
		colorAttention: colorAttention,
	}
}

func CreateAllBorders(
	parent *xwindow.Window, size int, colorNormal, colorActive, colorAttention uint32,
) []Decoration {
	pos := []Position{Left, Right, Top, Bottom}
	borders := make([]Decoration, 4)

	for i, p := range pos {
		borders[i] = CreateBorder(parent, p, size, colorNormal, colorActive, colorAttention)
	}

	return borders
}

func (b *Border) ApplyRect(config *WinConfig, rect xrect.Rect, f int) xrect.Rect {
	newRect := xrect.New(rect.Pieces())

	if config.Fullscreen {
		b.win.Unmap()
		return newRect
	}

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

	b.win.MROpt(f, x, y, w, h)
	b.win.Map()

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

func (b *Border) WidthNeeded(config *WinConfig, ) int {
	return b.sizeIfPos(config, Left, Right)
}

func (b *Border) HeightNeeded(config *WinConfig, ) int {
	return b.sizeIfPos(config, Top, Bottom)
}

func (b *Border) Left(config *WinConfig) int {
	return b.sizeIfPos(config, Left)
}

func (b *Border) Right(config *WinConfig) int {
	return b.sizeIfPos(config, Right)
}

func (b *Border) Top(config *WinConfig) int {
	return b.sizeIfPos(config, Top)
}

func (b *Border) Bottom(config *WinConfig) int {
	return b.sizeIfPos(config, Bottom)
}

func (b *Border) Active() {
	b.win.Change(xproto.CwBackPixel, b.colorActive)
	b.win.ClearAll()
}

func (b *Border) InActive() {
	b.win.Change(xproto.CwBackPixel, b.colorNormal)
	b.win.ClearAll()
}

func (b *Border) Attention() {
	b.win.Change(xproto.CwBackPixel, b.colorAttention)
	b.win.ClearAll()
}

func (b *Border) Destroy() {
	b.win.Destroy()
}

func (b *Border) sizeIfPos(config *WinConfig, pos ...Position) int {
	if config.Fullscreen {
		return 0
	}
	for _, p := range pos {
		if b.position == p {
			return b.size
		}
	}
	return 0
}
