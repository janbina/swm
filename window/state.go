package window

import (
	"github.com/janbina/swm/geometry"
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
	w.MoveResize(s.geom.PiecesTotal())
}
