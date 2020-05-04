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
	f := ConfigAll
	if s == StatePriorMaxVert {
		f = ConfigY | ConfigHeight
	} else if s == StatePriorMaxHorz {
		f = ConfigX | ConfigWidth
	}

	w.moveResizeInternal(false, g.X(), g.Y(), g.Width(), g.Height(), f)
}
