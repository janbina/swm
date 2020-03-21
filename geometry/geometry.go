package geometry

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xwindow"
)

type Geometry struct {
	x      int
	y      int
	w      int
	h      int
	border int
}

func (g *Geometry) X() int {
	return g.x
}

func (g *Geometry) Y() int {
	return g.y
}

func (g *Geometry) Width() int {
	return g.w
}

func (g *Geometry) Height() int {
	return g.h
}

func (g *Geometry) BorderWidth() int {
	return g.border
}

func (g *Geometry) TotalWidth() int {
	return g.w + g.border
}

func (g *Geometry) TotalHeight() int {
	return g.h + g.border
}

func (g *Geometry) SetX(v int) {
	g.x = v
}

func (g *Geometry) SetY(v int) {
	g.y = v
}

func (g *Geometry) SetWidth(v int) {
	g.w = v
}

func (g *Geometry) SetHeight(v int) {
	g.h = v
}

func (g *Geometry) AddX(v int) {
	g.x += v
}

func (g *Geometry) AddY(v int) {
	g.y += v
}

func (g *Geometry) AddWidth(v int) {
	g.w += v
}

func (g *Geometry) AddHeight(v int) {
	g.h += v
}

func (g *Geometry) Pieces() (int, int, int, int) {
	return g.X(), g.Y(), g.Width(), g.Height()
}

func Get(win *xwindow.Window) (*Geometry, error) {
	g, err := xproto.GetGeometry(win.X.Conn(), xproto.Drawable(win.Id)).Reply()
	if err != nil {
		return nil, err
	}
	return &Geometry{
		x:      int(g.X),
		y:      int(g.Y),
		w:      int(g.Width),
		h:      int(g.Height),
		border: int(g.BorderWidth),
	}, nil
}
