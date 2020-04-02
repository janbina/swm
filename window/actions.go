package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/janbina/swm/focus"
	"github.com/janbina/swm/heads"
	"github.com/janbina/swm/stack"
	"github.com/janbina/swm/util"
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
	w.parent.Move(x, y)
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
	w.parent.MoveResize(x, y, realWidth, realHeight)
	w.win.Resize(realWidth, realHeight)
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
	w.parent.Configure(flags, x, y, realWidth, realHeight, 0, 0)
	w.win.Configure(flags, 0, 0, realWidth, realHeight, 0, 0)
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
	winG, err := w.Geometry()
	if err != nil {
		log.Printf("Cannot get window geometry: %s", err)
	}
	g, err := heads.GetHeadForRectStruts(winG.RectTotal())
	if err != nil {
		log.Printf("Cannot get screen geometry: %s", err)
	}
	w.Configure(xproto.ConfigWindowY|xproto.ConfigWindowHeight, 0, g.Y(), 0, g.Height())
}

func (w *Window) UnMaximizeVert() {
	if !w.maxedVert {
		return
	}
	w.maxedVert = false
	w.removeStates("_NET_WM_STATE_MAXIMIZED_VERT", "MAXIMIZED")

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
	winG, err := w.Geometry()
	if err != nil {
		log.Printf("Cannot get window geometry: %s", err)
	}
	g, err := heads.GetHeadForRectStruts(winG.RectTotal())
	if err != nil {
		log.Printf("Cannot get screen geometry: %s", err)
	}
	w.Configure(xproto.ConfigWindowX|xproto.ConfigWindowWidth, g.X(), 0, g.Width(), 0)
}

func (w *Window) UnMaximizeHorz() {
	if !w.maxedHorz {
		return
	}
	w.maxedHorz = false
	w.removeStates("_NET_WM_STATE_MAXIMIZED_HORZ", "MAXIMIZED")

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
	w.removeStates("_NET_WM_STATE_MAXIMIZED_HORZ", "_NET_WM_STATE_MAXIMIZED_VERT", "MAXIMIZED")
}

func (w *Window) IconifyToggle() {
	if w.iconified {
		w.Show()
	} else {
		w.Hide()
	}
	focus.FocusLast()
}

func (w *Window) UnStackAbove() {
	w.layer = stack.LayerDefault
	w.Raise()
	w.removeStates("_NET_WM_STATE_ABOVE")
}

func (w *Window) StackAbove() {
	w.layer = stack.LayerAbove
	w.Raise()
	w.addStates("_NET_WM_STATE_ABOVE")
}

func (w *Window) StackAboveToggle() {
	if w.layer == stack.LayerAbove {
		w.UnStackAbove()
	} else {
		w.StackAbove()
	}
}

func (w *Window) UnStackBelow() {
	w.layer = stack.LayerDefault
	w.Raise()
	w.removeStates("_NET_WM_STATE_BELOW")
}

func (w *Window) StackBelow() {
	w.layer = stack.LayerBelow
	w.Raise()
	w.addStates("_NET_WM_STATE_BELOW")
}

func (w *Window) StackBelowToggle() {
	if w.layer == stack.LayerBelow {
		w.UnStackBelow()
	} else {
		w.StackBelow()
	}
}

func (w *Window) StopAttention() {
	w.demandsAttention = false
	_ = util.SetBorderColor(w.parent, borderColorInactive)
	w.removeStates("_NET_WM_STATE_DEMANDS_ATTENTION")
}

func (w *Window) StartAttention() {
	w.demandsAttention = true
	_ = util.SetBorderColor(w.parent, borderColorAttention)
	w.addStates("_NET_WM_STATE_DEMANDS_ATTENTION")
}

func (w *Window) ToggleAttention() {
	if w.demandsAttention {
		w.StopAttention()
	} else {
		w.StartAttention()
	}
}
