package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/janbina/swm/focus"
	"log"
)

type Directions struct {
	Left   int
	Right  int
	Bottom int
	Top    int
}

func (w *Window) Resize(d Directions) {
	g, _ := w.Geometry()
	x := g.X() + d.Left
	y := g.Y() + d.Top

	width := g.TotalWidth() + d.Right - d.Left
	height := g.TotalHeight() + d.Bottom - d.Top
	w.MoveResize(x, y, width, height)
}

func (w *Window) Move(x, y int) {
	w.UnsetMaximized()
	w.win.Move(x, y)
}

func (w *Window) MoveResize(x, y, width, height int) {
	g, _ := w.Geometry()
	realWidth := width - 2*g.BorderWidth()
	realHeight := height - 2*g.BorderWidth()

	if realWidth < int(w.normalHints.MinWidth) {
		realWidth = int(w.normalHints.MinWidth)
	}
	if realHeight < int(w.normalHints.MinHeight) {
		realHeight = int(w.normalHints.MinHeight)
	}
	w.UnsetMaximized()
	w.win.MoveResize(x, y, realWidth, realHeight)
}

func (w *Window) Configure(flags, x, y, width, height int) {
	g, _ := w.Geometry()
	realWidth := width - 2*g.BorderWidth()
	realHeight := height - 2*g.BorderWidth()

	if realWidth < int(w.normalHints.MinWidth) {
		realWidth = int(w.normalHints.MinWidth)
	}
	if realHeight < int(w.normalHints.MinHeight) {
		realHeight = int(w.normalHints.MinHeight)
	}
	w.win.Configure(flags, x, y, realWidth, realHeight, 0, 0)
}

func (w *Window) Maximize() {
	w.MaximizeHorz()
	w.MaximizeVert()
}

func (w *Window) UnMaximize() {
	w.UnMaximizeVert()
	w.UnMaximizeHorz()
}

func (w *Window) MaximizeToggle() {
	if w.maxedVert && w.maxedHorz {
		w.UnMaximize()
	} else {
		w.Maximize()
	}
}

func (w *Window) MaximizeVert() {
	if w.maxedVert {
		return
	}
	w.maxedVert = true
	w.addStates("_NET_WM_STATE_MAXIMIZED_VERT")

	w.SaveWindowState("prior_maxVert")
	g, _ := xwindow.New(w.win.X, w.win.X.RootWin()).Geometry() // TODO: get real geometry
	log.Printf("GEOM: %s", g)
	w.Configure(xproto.ConfigWindowY|xproto.ConfigWindowHeight, 0, g.Y(), 0, g.Height())
}

func (w *Window) UnMaximizeVert() {
	if !w.maxedVert {
		return
	}
	w.maxedVert = false
	w.removeStates("_NET_WM_STATE_MAXIMIZED_VERT")

	w.LoadWindowState("prior_maxVert")
}

func (w *Window) MaximizeVertToggle() {
	if w.maxedVert {
		w.UnMaximizeVert()
	} else {
		w.MaximizeVert()
	}
}

func (w *Window) MaximizeHorz() {
	if w.maxedHorz {
		return
	}
	w.maxedHorz = true
	w.addStates("_NET_WM_STATE_MAXIMIZED_HORZ")

	w.SaveWindowState("prior_maxHorz")
	g, _ := xwindow.New(w.win.X, w.win.X.RootWin()).Geometry() // TODO: get real geometry
	w.Configure(xproto.ConfigWindowX|xproto.ConfigWindowWidth, g.X(), 0, g.Width(), 0)
}

func (w *Window) UnMaximizeHorz() {
	if !w.maxedHorz {
		return
	}
	w.maxedHorz = false
	w.removeStates("_NET_WM_STATE_MAXIMIZED_HORZ")

	w.LoadWindowState("prior_maxHorz")
}

func (w *Window) MaximizeHorzToggle() {
	if w.maxedHorz {
		w.UnMaximizeHorz()
	} else {
		w.MaximizeHorz()
	}
}

func (w *Window) UnsetMaximized() {
	w.maxedVert = false
	w.maxedHorz = false
	w.removeStates("_NET_WM_STATE_MAXIMIZED_HORZ", "_NET_WM_STATE_MAXIMIZED_VERT")
}

func (w *Window) IconifyToggle() {
	if w.iconified {
		w.iconified = false
		w.Map()
		w.removeStates("_NET_WM_STATE_HIDDEN")
	} else {
		w.iconified = true
		w.Unmap()
		w.addStates("_NET_WM_STATE_HIDDEN")
	}
	focus.FocusLast()
}
