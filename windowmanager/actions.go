package windowmanager

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/janbina/swm/desktopmanager"
	"github.com/janbina/swm/focus"
	"github.com/janbina/swm/geometry"
	"github.com/janbina/swm/heads"
	"github.com/janbina/swm/stack"
	"github.com/janbina/swm/util"
	"github.com/janbina/swm/window"
)

var moveDragShortcut = "Mod1-1"
var resizeDragShortcut = "Mod1-3"

func FindWindowById(id xproto.Window) ManagedWindow {
	return managedWindows[id]
}

func getActiveWin() focus.FocusableWindow {
	return focus.Current()
}

func DoOnActiveWindow(f func(*window.Window)) {
	if w := getActiveWin(); w != nil {
		DoOnWindow(w.Id(), f)
	}
}

func DoOnWindow(id xproto.Window, f func(*window.Window)) {
	win := FindWindowById(id).(*window.Window)
	if win != nil {
		f(win)
	}
}

func ResizeActiveWindow(directions window.Directions) {
	if w := getActiveWin(); w != nil {
		ResizeWindow(w.Id(), directions)
	}
}

func ResizeWindow(id xproto.Window, directions window.Directions) {
	win := FindWindowById(id)
	if win == nil {
		return
	}
	win.Resize(directions)
}

func MoveActiveWindow(x, y int) {
	if w := getActiveWin(); w != nil {
		MoveWindow(w.Id(), x, y)
	}
}

func MoveWindow(id xproto.Window, x, y int) {
	win := FindWindowById(id)
	if win == nil {
		return
	}
	win.Move(x, y)
}

func MoveResizeActiveWindow(x, y, width, height int) {
	if w := getActiveWin(); w != nil {
		MoveResizeWindow(w.Id(), x, y, width, height)
	}
}

func MoveResizeWindow(id xproto.Window, x, y, width, height int) {
	win := FindWindowById(id)
	if win == nil {
		return
	}
	win.MoveResize(x, y, width, height)
}

func SetMoveDragShortcut(s string) error {
	if _, _, err := mousebind.ParseString(X, s); err != nil {
		return err
	}
	moveDragShortcut = s
	mouseShortcutsChanged()
	return nil
}

func SetResizeDragShortcut(s string) error {
	if _, _, err := mousebind.ParseString(X, s); err != nil {
		return err
	}
	resizeDragShortcut = s
	mouseShortcutsChanged()
	return nil
}

func GetCurrentScreenGeometry() (xrect.Rect, error) {
	if w := getActiveWin(); w != nil {
		return GetWindowScreenGeometry(w.Id())
	}
	return nil, fmt.Errorf("no active window")
}

func GetWindowScreenGeometry(id xproto.Window) (xrect.Rect, error) {
	winGeom, err := GetWindowGeometry(id)
	if err != nil {
		return nil, err
	}
	return heads.GetHeadForRect(winGeom.RectTotal())
}

func GetCurrentScreenGeometryStruts() (xrect.Rect, error) {
	if w := getActiveWin(); w != nil {
		return GetWindowScreenGeometryStruts(w.Id())
	}
	return nil, fmt.Errorf("no active window")
}

func GetWindowScreenGeometryStruts(id xproto.Window) (xrect.Rect, error) {
	winGeom, err := GetWindowGeometry(id)
	if err != nil {
		return nil, err
	}
	return heads.GetHeadForRectStruts(winGeom.RectTotal())
}

func GetActiveWindowGeometry() (*geometry.Geometry, error) {
	if w := getActiveWin(); w != nil {
		return GetWindowGeometry(w.Id())
	}
	return nil, fmt.Errorf("no active window")
}

func GetWindowGeometry(id xproto.Window) (*geometry.Geometry, error) {
	win := FindWindowById(id)
	if win == nil {
		return nil, fmt.Errorf("cannot find window with id %d", id)
	}
	return win.Geometry()
}

func setNumberOfDesktops(num int) {
	changes := desktopmanager.SetNumberOfDesktops(num)
	applyChanges(changes)
	setWorkArea(desktopmanager.GetNumDesktops())
	focus.FocusLast()
}

func switchToDesktop(index int) {
	changes := desktopmanager.SwitchToDesktop(index)
	applyChanges(changes)
	focus.FocusLast()
}

func applyChanges(changes *desktopmanager.Changes) {
	for _, w := range changes.Invisible {
		win := managedWindows[w]
		if win == nil {
			panic("This shouldnt happen anymore")
		}
		win.Unmap()
	}
	for _, w := range changes.Visible {
		win := managedWindows[w]
		if win == nil {
			panic("This shouldnt happen anymore")
		}
		if !win.IsHidden() {
			win.Map()
		}
	}
}

func mouseShortcutsChanged() {
	for _, win := range managedWindows {
		win.SetupMouseEvents(moveDragShortcut, resizeDragShortcut)
	}
}

func CycleWin() {
	cycleState--
	if win, ok := focus.CyclingFocus(cycleState).(*window.Window); ok {
		stack.TmpRaise(win)
	}
}

func CycleWinRev() {
	cycleState++
	if win, ok := focus.CyclingFocus(cycleState).(*window.Window); ok {
		stack.TmpRaise(win)
	}
}

func CycleWinEnd() {
	cycleState = 0
	if win, ok := focus.CyclingEnded().(*window.Window); ok {
		win.RemoveTmpDeiconified()
		win.Raise()
	}
}

func MoveWindowToDesktop(w *window.Window, desktop int) {
	changes := desktopmanager.MoveWindowToDesktop(w.Id(), desktop)
	applyChanges(changes)
	focus.FocusLast()
}

func UnstickWindow(w *window.Window) {
	if desktopmanager.IsWinSticky(w.Id()) {
		MoveWindowToDesktop(w, desktopmanager.GetCurrentDesktop())
		w.RemoveStates("_NET_WM_STATE_STICKY")
	}
}

func StickWindow(w *window.Window) {
	MoveWindowToDesktop(w, desktopmanager.StickyDesktop)
	w.AddStates("_NET_WM_STATE_STICKY")
}

func ToggleWindowSticky(w *window.Window) {
	if desktopmanager.IsWinSticky(w.Id()) {
		UnstickWindow(w)
	} else {
		StickWindow(w)
	}
}

func BeginMouseMoveFromPointer() error {
	p, err := util.QueryPointerClient(X)
	if err != nil {
		return fmt.Errorf("no client window underneath the pointer")
	}
	win := FindWindowById(p.Win)
	if win == nil {
		return fmt.Errorf("no client window underneath the pointer")
	}
	win.(*window.Window).DragMoveBegin(int16(p.X), int16(p.Y))
	return nil
}

func BeginMouseResizeFromPointer() error {
	p, err := util.QueryPointerClient(X)
	if err != nil {
		return fmt.Errorf("no client window underneath the pointer")
	}
	win := FindWindowById(p.Win)
	if win == nil {
		return fmt.Errorf("no client window underneath the pointer")
	}
	win.(*window.Window).DragResizeBeginEvent(int16(p.X), int16(p.Y), int16(p.WinX), int16(p.WinY))
	return nil
}
