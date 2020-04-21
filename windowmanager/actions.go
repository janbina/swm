package windowmanager

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/janbina/swm/desktopmanager"
	"github.com/janbina/swm/focus"
	"github.com/janbina/swm/heads"
	"github.com/janbina/swm/stack"
	"github.com/janbina/swm/util"
	"github.com/janbina/swm/window"
)

var moveDragShortcut = "Mod1-1"
var resizeDragShortcut = "Mod1-3"

func getActiveWindow() focus.FocusableWindow {
	return focus.Current()
}

func GetWindowById(id int) (ManagedWindow, error) {
	if id == 0 {
		if active := getActiveWindow(); active == nil {
			return nil, fmt.Errorf("cannot get active window")
		} else {
			id = int(active.Id())
		}
	}
	if win := managedWindows[xproto.Window(id)]; win == nil {
		return nil, fmt.Errorf("cannot find window with id %d", id)
	} else {
		return win, nil
	}
}

func doOnWindow(id int, action func(win ManagedWindow)) error {
	win, err := GetWindowById(id)
	if err != nil {
		return err
	}
	action(win)
	return nil
}

func MoveWindow(id int, x, y int) error {
	return doOnWindow(id, func(win ManagedWindow) {
		win.Move(x, y)
	})
}

func MoveResizeWindow(id int, x, y, width, height int) error {
	return doOnWindow(id, func(win ManagedWindow) {
		win.MoveResize(true, x, y, width, height)
	})
}

func GetWindowScreenGeometry(id int) (xrect.Rect, error) {
	winGeom, err := GetWindowGeometry(id)
	if err != nil {
		return nil, err
	}
	return heads.GetHeadForRect(winGeom)
}

func GetWindowScreenGeometryStruts(id int) (xrect.Rect, error) {
	winGeom, err := GetWindowGeometry(id)
	if err != nil {
		return nil, err
	}
	return heads.GetHeadForRectStruts(winGeom)
}

func GetWindowGeometry(id int) (xrect.Rect, error) {
	win, err := GetWindowById(id)
	if err != nil {
		return nil, err
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

func switchToWindowDesktop(win xproto.Window) {
	if !desktopmanager.IsWinDesktopVisible(win) {
		desktopmanager.SwitchToDesktop(desktopmanager.GetWinDesktop(win))
	}
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

func mouseShortcutsChanged() {
	for _, win := range managedWindows {
		win.SetupMouseEvents(moveDragShortcut, resizeDragShortcut)
	}
}

func BeginMouseMoveFromPointer() error {
	p, err := util.QueryPointerClient(X)
	if err != nil {
		return fmt.Errorf("no client window underneath the pointer")
	}
	win, err := GetWindowById(int(p.Win))
	if err != nil {
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
	win, err := GetWindowById(int(p.Win))
	if err != nil {
		return fmt.Errorf("no client window underneath the pointer")
	}
	win.(*window.Window).DragResizeBeginEvent(int16(p.X), int16(p.Y), int16(p.WinX), int16(p.WinY))
	return nil
}
