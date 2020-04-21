package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/janbina/swm/decoration"
	"github.com/janbina/swm/heads"
	"github.com/janbina/swm/stack"
	"github.com/janbina/swm/util"
	"log"
)

const (
	ConfigX        = xproto.ConfigWindowX
	ConfigY        = xproto.ConfigWindowY
	ConfigWidth    = xproto.ConfigWindowWidth
	ConfigHeight   = xproto.ConfigWindowHeight
	ConfigPosition = ConfigX | ConfigY
	ConfigSize     = ConfigWidth | ConfigHeight
	ConfigAll      = ConfigPosition | ConfigSize
)

func (w *Window) Move(x, y int) {
	w.MoveResize(x, y, 0, 0, ConfigPosition)
}

// single function for all moving and/or resizing, which also automatically cancel fullscreen and maximized state
func (w *Window) MoveResize(x, y, width, height int, flags ...int) {
	w.UnFullscreen()
	w.UnMaximizeVert()
	w.UnMaximizeHorz()
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
		w.sendConfigureNotify()
	} else {
		extents := w.GetFrameExtents()
		parentWidth := width + extents.Left + extents.Right
		parentHeight := height + extents.Top + extents.Bottom

		w.parent.MROpt(f, x, y, parentWidth, parentHeight)

		rect := xrect.New(0, 0, parentWidth, parentHeight)
		newRect := w.decorations.ApplyRect(&decoration.WinConfig{Fullscreen:w.fullscreen}, rect)


		//g, _ := w.Geometry()
		//realWidth := width - 2*g.BorderWidth()
		//realHeight := height - 2*g.BorderWidth()
		//
		//if realWidth < int(w.normalHints.MinWidth) {
		//	realWidth = int(w.normalHints.MinWidth)
		//}
		//if realHeight < int(w.normalHints.MinHeight) {
		//	realHeight = int(w.normalHints.MinHeight)
		//}

		w.win.MROpt(f, newRect.X(), newRect.Y(), newRect.Width(), newRect.Height())
		w.sendConfigureNotify()
	}
}

func (w *Window) MaximizeVert() {
	if w.maxedVert || w.fullscreen {
		return
	}
	w.maxedVert = true
	w.AddStates("_NET_WM_STATE_MAXIMIZED_VERT")

	w.SaveWindowState(StatePriorMaxVert)
	winG, err := w.Geometry()
	if err != nil {
		log.Printf("Cannot get window geometry: %s", err)
	}
	g, err := heads.GetHeadForRectStruts(winG)
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
	w.RemoveStates("_NET_WM_STATE_MAXIMIZED_VERT")

	w.LoadWindowState(StatePriorMaxVert)
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

	w.SaveWindowState(StatePriorMaxHorz)
	winG, err := w.Geometry()
	if err != nil {
		log.Printf("Cannot get window geometry: %s", err)
	}
	g, err := heads.GetHeadForRectStruts(winG)
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
	w.RemoveStates("_NET_WM_STATE_MAXIMIZED_HORZ")

	w.LoadWindowState(StatePriorMaxHorz)
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
	w.decorations.InActive()
	w.RemoveStates("_NET_WM_STATE_DEMANDS_ATTENTION")
}

func (w *Window) StartAttention() {
	w.demandsAttention = true
	w.decorations.Attention()
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

	w.LoadWindowState(StatePriorFullscreen)

	w.layer = stack.LayerDefault
	stack.ReStack()
}

func (w *Window) Fullscreen() {
	if w.fullscreen {
		return
	}
	w.fullscreen = true
	w.AddStates("_NET_WM_STATE_FULLSCREEN")

	w.SaveWindowState(StatePriorFullscreen)
	winG, err := w.Geometry()
	if err != nil {
		log.Printf("Cannot get window geometry: %s", err)
	}
	g, err := heads.GetHeadForRect(winG)
	if err != nil {
		log.Printf("Cannot get screen geometry: %s", err)
	}
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

func (w *Window) UnSkipTaskbar() {
	w.skipTaskbar = false
	w.RemoveStates("_NET_WM_STATE_SKIP_TASKBAR")
}

func (w *Window) SkipTaskbar() {
	w.skipTaskbar = true
	w.AddStates("_NET_WM_STATE_SKIP_TASKBAR")
}

func (w *Window) ToggleSkipTaskbar() {
	if w.skipTaskbar {
		w.UnSkipTaskbar()
	} else {
		w.SkipTaskbar()
	}
}

func (w *Window) UnSkipPager() {
	w.skipPager = true
	w.RemoveStates("_NET_WM_STATE_SKIP_PAGER")
}

func (w *Window) SkipPager() {
	w.skipPager = true
	w.AddStates("_NET_WM_STATE_SKIP_PAGER")
}

func (w *Window) ToggleSkipPager() {
	if w.skipPager {
		w.UnSkipPager()
	} else {
		w.SkipPager()
	}
}

func (w *Window) ConfigureRequest(e xevent.ConfigureRequestEvent) {
	log.Printf("Window configure request: %s", e)
	flags := int(e.ValueMask)
	x, y, width, height := int(e.X), int(e.Y), int(e.Width), int(e.Height)

	if flags&ConfigAll != 0 {
		w.MoveResize(x, y, width, height, flags)
	}
}

// RootGeometryChanged moves window based on changes to root geometry
// New monitors might have been added/removed and resolution could have changed
// We unfullscreen and unmaximize window, so it restores its original geometry,
// than we check if window overlaps with any monitor and if not, move it so it does
// and finally restore maximized and fullscreen states
func (w *Window) RootGeometryChanged() {
	maxedVert, maxedHorz, fullscreen := w.maxedVert, w.maxedHorz, w.fullscreen
	w.UnFullscreen()
	w.UnMaximizeVert()
	w.UnMaximizeHorz()

	g, _ := w.Geometry()

	dX, dY := util.MinMovement(g, heads.HeadsStruts, 50)
	flags := 0
	if dX != 0 {
		flags |= ConfigX
	}
	if dY != 0 {
		flags |= ConfigY
	}
	if flags != 0 {
		w.MoveResize(g.X()+dX, g.Y()+dY, 0, 0, flags)
	}

	if maxedVert {
		w.MaximizeVert()
	}
	if maxedHorz {
		w.MaximizeHorz()
	}
	if fullscreen {
		w.Fullscreen()
	}
}

func (w *Window) sendConfigureNotify() {
	if g, err := w.Geometry(); err == nil {
		e := xproto.ConfigureNotifyEvent{
			Event:            w.win.Id,
			Window:           w.win.Id,
			AboveSibling:     0,
			X:                int16(g.X()) + 1,
			Y:                int16(g.Y()) + 1,
			Width:            uint16(g.Width()),
			Height:           uint16(g.Height()),
			BorderWidth:      0,
			OverrideRedirect: false,
		}
		xproto.SendEvent(
			w.win.X.Conn(),
			false,
			w.win.Id,
			xproto.EventMaskStructureNotify,
			string(e.Bytes()),
		)
	}
}

func (w *Window) updateFrameExtents() {
	_ = ewmh.FrameExtentsSet(w.win.X, w.win.Id, w.GetFrameExtents())
}

func (w *Window) GetFrameExtents() *ewmh.FrameExtents {
	config := &decoration.WinConfig{Fullscreen: w.fullscreen}
	return &ewmh.FrameExtents{
		Left:   w.decorations.Left(config),
		Right:  w.decorations.Right(config),
		Top:    w.decorations.Top(config),
		Bottom: w.decorations.Bottom(config),
	}
}
