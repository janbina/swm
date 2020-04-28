package windowmanager

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/janbina/swm/config"
	"github.com/janbina/swm/groupmanager"
	"github.com/janbina/swm/focus"
	"github.com/janbina/swm/heads"
	"github.com/janbina/swm/stack"
	"github.com/janbina/swm/util"
	"github.com/janbina/swm/window"
)

func getActiveWindow() focus.FocusableWindow {
	return focus.Current()
}

func GetWindowById(id int) (*window.Window, error) {
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

func doOnWindow(id int, action func(win *window.Window)) error {
	win, err := GetWindowById(id)
	if err != nil {
		return err
	}
	action(win)
	return nil
}

func MoveWindow(id int, x, y int) error {
	return doOnWindow(id, func(win *window.Window) {
		win.Move(x, y)
	})
}

func MoveResizeWindow(id int, x, y, width, height int) error {
	return doOnWindow(id, func(win *window.Window) {
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

func SetMoveDragShortcut(s string) error {
	if _, _, err := mousebind.ParseString(X, s); err != nil {
		return err
	}
	config.MoveDragShortcut = s
	mouseShortcutsChanged()
	return nil
}

func SetResizeDragShortcut(s string) error {
	if _, _, err := mousebind.ParseString(X, s); err != nil {
		return err
	}
	config.ResizeDragShortcut = s
	mouseShortcutsChanged()
	return nil
}

func mouseShortcutsChanged() {
	for _, win := range managedWindows {
		win.SetupMouseEvents()
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
	win.DragMoveBegin(int16(p.X), int16(p.Y))
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
	win.DragResizeBeginEvent(int16(p.X), int16(p.Y), int16(p.WinX), int16(p.WinY))
	return nil
}

// GROUPS

func SetGroupForWindow(w xproto.Window, desktop int) {
	changes := groupmanager.SetGroupForWindow(w, desktop)
	applyChanges(changes)
	focus.FocusLast()
}

func SetGroupForActiveWindow(group int) error {
	active := getActiveWindow()
	if active == nil {
		return fmt.Errorf("cannot get active window")
	}
	SetGroupForWindow(active.Id(), group)
	return nil
}

func setNumberOfDesktops(num int) {
	changes := groupmanager.SetNumberOfGroups(num)
	applyChanges(changes)
	setWorkArea(groupmanager.GetNumGroups())
	focus.FocusLast()
}

func switchToDesktop(index int) {
	ShowGroupOnly(index)
}

func showWindowGroup(win xproto.Window) {
	ShowGroup(groupmanager.GetWinGroup(win))
}

func ToggleGroupVisibility(group int) {
	changes := groupmanager.ToggleGroupVisibility(group)
	applyChanges(changes)
	raiseChanges(changes)
	focus.FocusLastWithPreference(func(win xproto.Window) bool {
		return groupmanager.GetWinGroup(win) == group
	})
}

func ShowGroupOnly(group int) {
	changes := groupmanager.ShowGroupOnly(group)
	applyChanges(changes)
	focus.FocusLast()
}

func ShowGroup(group int) {
	changes := groupmanager.ShowGroup(group)
	applyChanges(changes)
	raiseChanges(changes)
	focus.FocusLastWithPreference(func(win xproto.Window) bool {
		return groupmanager.GetWinGroup(win) == group
	})
}

func HideGroup(group int) {
	changes := groupmanager.HideGroup(group)
	applyChanges(changes)
	focus.FocusLast()
}

func applyChanges(changes *groupmanager.Changes) {
	if changes == nil {
		return
	}
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

func raiseChanges(changes *groupmanager.Changes) {
	if changes == nil {
		return
	}
	wins := make([]stack.StackingWindow, 0, len(changes.Visible))
	for _, id := range changes.Visible {
		if win := managedWindows[id]; win != nil {
			wins = append(wins, win)
		}
	}
	stack.RaiseMulti(wins)
}
