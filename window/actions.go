package window

import (
	"github.com/BurntSushi/xgb/xproto"
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

const (
	ConfigX        = xproto.ConfigWindowX
	ConfigY        = xproto.ConfigWindowY
	ConfigWidth    = xproto.ConfigWindowWidth
	ConfigHeight   = xproto.ConfigWindowHeight
	ConfigPosition = ConfigX | ConfigY
	ConfigSize     = ConfigWidth | ConfigHeight
	ConfigAll      = ConfigPosition | ConfigSize
)

func (w *Window) Resize(d Directions) {
	g, _ := w.Geometry()
	x := g.X() + d.Left
	y := g.Y() + d.Top

	width := g.TotalWidth() + d.Right - d.Left
	height := g.TotalHeight() + d.Bottom - d.Top
	w.MoveResize(x, y, width, height)
}

func (w *Window) Move(x, y int) {
	w.MoveResize(x, y, 0, 0, ConfigPosition)
}

// single function for all moving and/or resizing, which also automatically cancel fullscreen and maximized state
func (w *Window) MoveResize(x, y, width, height int, flags ...int) {
	w.UnFullscreen()
	w.UnMaximize()
	w.moveResizeInternal(x, y, width, height, flags...)
}

// single function for all moving and/or resizing, without any side effects
// call this if you are sure you dont want any side effects (canceled fullscreen and maximized state), otherwise,
// use MoveResize() or Move()
func (w *Window) moveResizeInternal(x, y, width, height int, flags ...int) {
	f := 0
	for _, flag := range flags {
		f |= flag
	}
	if len(flags) == 0 {
		f = ConfigAll
	}
	onlyMove := f&ConfigSize == 0
	if onlyMove {
		w.parent.Move(x, y)
	} else {
		g, _ := w.Geometry()
		realWidth := width - 2*g.BorderWidth()
		realHeight := height - 2*g.BorderWidth()

		if realWidth < int(w.normalHints.MinWidth) {
			realWidth = int(w.normalHints.MinWidth)
		}
		if realHeight < int(w.normalHints.MinHeight) {
			realHeight = int(w.normalHints.MinHeight)
		}
		w.parent.MROpt(f, x, y, realWidth, realHeight)
		w.win.MROpt(f, 0, 0, realWidth, realHeight)
	}
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
	if w.maxedVert || w.fullscreen {
		return
	}
	w.maxedVert = true
	w.AddStates("_NET_WM_STATE_MAXIMIZED_VERT")

	w.SaveWindowState("prior_maxVert")
	winG, err := w.Geometry()
	if err != nil {
		log.Printf("Cannot get window geometry: %s", err)
	}
	g, err := heads.GetHeadForRectStruts(winG.RectTotal())
	if err != nil {
		log.Printf("Cannot get screen geometry: %s", err)
	}
	w.moveResizeInternal(0, g.Y(), 0, g.Height(), ConfigY, ConfigHeight)
}

func (w *Window) UnMaximizeVert() {
	if !w.maxedVert || w.fullscreen {
		return
	}
	w.maxedVert = false
	w.RemoveStates("_NET_WM_STATE_MAXIMIZED_VERT", "MAXIMIZED")

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
	if w.maxedHorz || w.fullscreen {
		return
	}
	w.maxedHorz = true
	w.AddStates("_NET_WM_STATE_MAXIMIZED_HORZ")

	w.SaveWindowState("prior_maxHorz")
	winG, err := w.Geometry()
	if err != nil {
		log.Printf("Cannot get window geometry: %s", err)
	}
	g, err := heads.GetHeadForRectStruts(winG.RectTotal())
	if err != nil {
		log.Printf("Cannot get screen geometry: %s", err)
	}
	w.moveResizeInternal(g.X(), 0, g.Width(), 0, ConfigX, ConfigWidth)
}

func (w *Window) UnMaximizeHorz() {
	if !w.maxedHorz || w.fullscreen {
		return
	}
	w.maxedHorz = false
	w.RemoveStates("_NET_WM_STATE_MAXIMIZED_HORZ", "MAXIMIZED")

	w.LoadWindowState("prior_maxHorz")
}

func (w *Window) MaximizeHorzToggle() {
	if w.maxedHorz {
		w.UnMaximizeHorz()
	} else {
		w.MaximizeHorz()
	}
}

func (w *Window) IconifyToggle() {
	if w.iconified {
		w.DeIconify()
	} else {
		w.Iconify()
	}
}

func (w *Window) UnStackAbove() {
	w.layer = stack.LayerDefault
	w.Raise()
	w.RemoveStates("_NET_WM_STATE_ABOVE")
}

func (w *Window) StackAbove() {
	w.layer = stack.LayerAbove
	w.Raise()
	w.AddStates("_NET_WM_STATE_ABOVE")
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
	w.RemoveStates("_NET_WM_STATE_BELOW")
}

func (w *Window) StackBelow() {
	w.layer = stack.LayerBelow
	w.Raise()
	w.AddStates("_NET_WM_STATE_BELOW")
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
	w.RemoveStates("_NET_WM_STATE_DEMANDS_ATTENTION")
}

func (w *Window) StartAttention() {
	w.demandsAttention = true
	_ = util.SetBorderColor(w.parent, borderColorAttention)
	w.AddStates("_NET_WM_STATE_DEMANDS_ATTENTION")
}

func (w *Window) ToggleAttention() {
	if w.demandsAttention {
		w.StopAttention()
	} else {
		w.StartAttention()
	}
}

func (w *Window) UnFullscreen() {
	if !w.fullscreen {
		return
	}
	w.fullscreen = false
	w.RemoveStates("_NET_WM_STATE_FULLSCREEN")

	w.LoadWindowState("prior_fullscreen")

	w.layer = stack.LayerDefault
	stack.ReStack()
}

func (w *Window) Fullscreen() {
	if w.fullscreen {
		return
	}
	w.fullscreen = true
	w.AddStates("_NET_WM_STATE_FULLSCREEN")

	w.SaveWindowState("prior_fullscreen")
	winG, err := w.Geometry()
	if err != nil {
		log.Printf("Cannot get window geometry: %s", err)
	}
	g, err := heads.GetHeadForRect(winG.RectTotal())
	if err != nil {
		log.Printf("Cannot get screen geometry: %s", err)
	}
	util.SetBorderWidth(w.parent, 0)
	w.moveResizeInternal(g.X(), g.Y(), g.Width(), g.Height())

	w.layer = stack.LayerFullscreen
	stack.ReStack()
}

func (w *Window) FullscreenToggle() {
	if w.fullscreen {
		w.UnFullscreen()
	} else {
		w.Fullscreen()
	}
}
