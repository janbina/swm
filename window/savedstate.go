package window

import (
	"github.com/janbina/swm/geometry"
	"github.com/janbina/swm/util"
)

type windowState struct {
	geom geometry.Geometry
}

func (w *Window) SaveWindowState(name string) {
	g, _ := w.Geometry()
	w.savedStates[name] = windowState{geom: *g}
}

func (w *Window) LoadWindowState(name string) {
	s, ok := w.savedStates[name]
	if !ok {
		return
	}
	w.moveResizeInternal(s.geom.PiecesTotal())
	util.SetBorderWidth(w.parent, uint32(s.geom.BorderWidth()))
	w.setFrameExtents(s.geom.BorderWidth())
}
