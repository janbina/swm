package window

import (
	"github.com/BurntSushi/xgbutil/xrect"
)

type state uint

const (
	StatePriorMaxVert state = iota
	StatePriorMaxHorz
	StatePriorFullscreen
)

type windowState struct {
	geom xrect.Rect
}

func (w *Window) SaveWindowState(s state) {
	g, _ := w.Geometry()
	w.savedStates[s] = windowState{geom: g}
}

func (w *Window) LoadWindowState(s state) {
	ws, ok := w.savedStates[s]
	if !ok {
		return
	}
	g := ws.geom
	if s == StatePriorMaxVert {
		w.moveResizeInternal(0, g.Y(), 0, g.Height(), ConfigY, ConfigHeight)
	} else if s == StatePriorMaxHorz {
		w.moveResizeInternal(g.X(), 0, g.Width(), 0, ConfigX, ConfigWidth)
	} else {
		w.moveResizeInternal(g.Pieces())
	}
}
