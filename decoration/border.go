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
	win      *xwindow.Window
	config   *BorderConfig
}

type BorderConfig struct {
	Size           int
	ColorNormal    uint32
	ColorActive    uint32
	ColorAttention uint32
}

func CreateBorder(parent *xwindow.Window, position Position, config *BorderConfig) Decoration {
	X := parent.X

	win, _ := xwindow.Create(X, parent.Id)
	win.Change(xproto.CwBackPixel, config.ColorNormal)

	return &Border{
		position: position,
		win:      win,
		config:   config,
	}
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
		w = b.config.Size
		h = rect.Height()
	case Top:
		x = rect.X()
		y = rect.Y()
		w = rect.Width()
		h = b.config.Size
	case Right:
		x = rect.X() + rect.Width() - b.config.Size
		y = rect.Y()
		w = b.config.Size
		h = rect.Height()
	case Bottom:
		x = rect.X()
		y = rect.Y() + rect.Height() - b.config.Size
		w = rect.Width()
		h = b.config.Size
	}

	b.win.MROpt(f, x, y, w, h)
	b.win.Map()

	switch b.position {
	case Left:
		newRect.XSet(newRect.X() + b.config.Size)
		newRect.WidthSet(newRect.Width() - b.config.Size)
	case Top:
		newRect.YSet(newRect.Y() + b.config.Size)
		newRect.HeightSet(newRect.Height() - b.config.Size)
	case Right:
		newRect.WidthSet(newRect.Width() - b.config.Size)
	case Bottom:
		newRect.HeightSet(newRect.Height() - b.config.Size)
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
	b.win.Change(xproto.CwBackPixel, b.config.ColorActive)
	b.win.ClearAll()
}

func (b *Border) InActive() {
	b.win.Change(xproto.CwBackPixel, b.config.ColorNormal)
	b.win.ClearAll()
}

func (b *Border) Attention() {
	b.win.Change(xproto.CwBackPixel, b.config.ColorAttention)
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
			return b.config.Size
		}
	}
	return 0
}
