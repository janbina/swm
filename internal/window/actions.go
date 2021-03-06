package window

import (
	"log"
	"math"
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/janbina/swm/internal/config"
	"github.com/janbina/swm/internal/decoration"
	"github.com/janbina/swm/internal/heads"
	"github.com/janbina/swm/internal/stack"
	"github.com/janbina/swm/internal/util"
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
	w.MoveResize(false, x, y, 0, 0, ConfigPosition)
}

func (w *Window) MoveResizeWinSize(validate bool, x, y, width, height int, flags ...int) {
	w.UnFullscreen()
	w.UnMaximizeVert()
	w.UnMaximizeHorz()
	e := w.GetFrameExtents()
	w.moveResizeInternal(validate, x, y, width+e.Left+e.Right, height+e.Top+e.Bottom, flags...)
}

// single function for all moving and/or resizing, which also automatically cancel fullscreen and maximized state
func (w *Window) MoveResize(validate bool, x, y, width, height int, flags ...int) {
	w.UnFullscreen()
	w.UnMaximizeVert()
	w.UnMaximizeHorz()
	w.moveResizeInternal(validate, x, y, width, height, flags...)
}

// single function for all moving and/or resizing, without any side effects
// call this if you are sure you dont want any side effects (canceled fullscreen and maximized state), otherwise,
// use MoveResize() or Move()
func (w *Window) moveResizeInternal(validate bool, x, y, width, height int, flags ...int) {
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
		innerWidth := width - extents.Left - extents.Right
		innerHeight := height - extents.Top - extents.Bottom

		if validate {
			innerWidth = int(w.ValidateWidth(uint(innerWidth)))
			innerHeight = int(w.ValidateHeight(uint(innerHeight)))
		}

		parentWidth := innerWidth + extents.Left + extents.Right
		parentHeight := innerHeight + extents.Top + extents.Bottom

		w.parent.MROpt(f, x, y, parentWidth, parentHeight)

		rect := xrect.New(0, 0, parentWidth, parentHeight)
		newRect := w.decorations.ApplyRect(&decoration.WinConfig{Fullscreen: w.fullscreen}, rect, f)

		if newRect.Width() != innerWidth || newRect.Height() != innerHeight {
			log.Printf("Bad window size")
		}

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
	w.moveResizeInternal(false, 0, g.Y(), 0, g.Height(), ConfigY, ConfigHeight)
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
	w.moveResizeInternal(false, g.X(), 0, g.Width(), 0, ConfigX, ConfigWidth)
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
	w.updateFrameExtents()

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
	w.moveResizeInternal(false, g.X(), g.Y(), g.Width(), g.Height())
	w.updateFrameExtents()

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
		w.MoveResizeWinSize(true, x, y, width, height, flags)
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
		w.MoveResize(true, g.X()+dX, g.Y()+dY, 0, 0, flags)
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
	e := w.GetFrameExtents()
	if g, err := w.Geometry(); err == nil {
		e := xproto.ConfigureNotifyEvent{
			Event:            w.win.Id,
			Window:           w.win.Id,
			AboveSibling:     0,
			X:                int16(g.X() + e.Left),
			Y:                int16(g.Y() + e.Top),
			Width:            uint16(g.Width() - e.Left - e.Right),
			Height:           uint16(g.Height() - e.Top - e.Bottom),
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

func (w *Window) ValidateHeight(height uint) uint {
	h := w.normalHints
	return w.validateSize(height, h.MinHeight, h.MaxHeight, h.BaseHeight, h.HeightInc)
}

func (w *Window) ValidateWidth(width uint) uint {
	h := w.normalHints
	return w.validateSize(width, h.MinWidth, h.MaxWidth, h.BaseWidth, h.WidthInc)
}

func (w *Window) validateSize(size, min, max, base, inc uint) uint {
	hints := w.normalHints

	if !hasFlag(hints, icccm.SizeHintPMinSize) && hasFlag(hints, icccm.SizeHintPBaseSize) {
		min = base
	}
	if !hasFlag(hints, icccm.SizeHintPBaseSize) && hasFlag(hints, icccm.SizeHintPMinSize) {
		base = min
	}
	hasMin := hasFlag(hints, icccm.SizeHintPMinSize) || hasFlag(hints, icccm.SizeHintPBaseSize)
	hasBase := hasMin

	if size < min && hasMin {
		return min
	}
	if size > max && hasFlag(hints, icccm.SizeHintPMaxSize) {
		return max
	}
	if inc > 1 && hasFlag(hints, icccm.SizeHintPResizeInc) && hasBase {
		// size = base + (i * inc)
		rem := size - base
		i := uint(math.Round(float64(rem) / float64(inc)))

		return base + i*inc
	}

	return size
}

func hasFlag(hints *icccm.NormalHints, flag uint) bool {
	return hints.Flags&flag > 0
}

func (w *Window) ShowInfoBox(text string, duration time.Duration) {
	if w.infoTimer != nil {
		if w.infoTimer.Stop() {
			w.infoTimer.Reset(0)
		}
	}

	textBox, err := util.CreateTextBox(
		w.win.X, text, 16, 5,
		config.InfoBoxBgColor,
		config.InfoBoxTextColor,
	)

	if err != nil {
		log.Printf("Cannot show info box: %s", err)
		return
	}

	w.infoWin.MoveResize(10, 10, textBox.Rect.Dx(), textBox.Rect.Dy())

	err = textBox.XSurfaceSet(w.infoWin.Id)
	if err != nil {
		log.Printf("Cannot set surface: %s", err)
		return
	}
	textBox.XDraw()
	textBox.XPaint(w.infoWin.Id)

	w.infoWin.Map()

	w.infoTimer = time.NewTimer(duration)
	go func() {
		<-w.infoTimer.C
		textBox.Destroy()
		w.HideInfoBox()
	}()
}

func (w *Window) HideInfoBox() {
	w.infoWin.Unmap()
}
