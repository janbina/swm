package windowmanager

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/janbina/swm/focus"
	"github.com/janbina/swm/geometry"
	"github.com/janbina/swm/window"
	"log"
)

var moveDragShortcut = "Mod1-1"
var resizeDragShortcut = "Mod1-3"

func FindWindowById(id uint32) ManagedWindow {
	return managedWindows[xproto.Window(id)]
}

func getActiveWin() focus.FocusableWindow {
	return focus.Current()
}

func DestroyActiveWindow() {
	if w := getActiveWin(); w != nil {
		DestroyWindow(w.Id())
	}
}

func DestroyWindow(id uint32) {
	win := FindWindowById(id)
	if win == nil {
		return
	}
	log.Printf("Destroy win %d", id)
	win.Destroy()
}

func ResizeActiveWindow(directions window.Directions) {
	if w := getActiveWin(); w != nil {
		ResizeWindow(w.Id(), directions)
	}
}

func ResizeWindow(id uint32, directions window.Directions) {
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

func MoveWindow(id uint32, x, y int) {
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

func MoveResizeWindow(id uint32, x, y, width, height int) {
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

func GetCurrentScreenGeometry() xrect.Rect {
	return Heads[0]
}

func GetActiveWindowGeometry() (*geometry.Geometry, error) {
	if w := getActiveWin(); w != nil {
		return GetWindowGeometry(w.Id())
	}
	return nil, fmt.Errorf("no active window")
}

func GetWindowGeometry(id uint32) (*geometry.Geometry, error) {
	win := FindWindowById(id)
	if win == nil {
		return nil, fmt.Errorf("cannot find window with id %d", id)
	}
	return win.Geometry()
}

func setNumberOfDesktops(num int) {
	if num < 1 {
		num = 1
	}
	currentNum := len(desktops)
	newLast := num - 1
	if num < currentNum {
		for i := num; i < currentNum; i++ {
			for _, x := range desktopToWins[i] {
				ewmh.WmDesktopSet(X, x, uint(newLast))
			}
			desktopToWins[newLast] = append(desktopToWins[newLast], desktopToWins[i]...)
		}
		desktops = desktops[:num]
		setDesktops()
		if currentDesktop > newLast {
			switchToDesktop(newLast)
		}
	} else if num > currentNum {
		desktops = append(desktops, getDesktopNames(currentNum, newLast)...)
		setDesktops()
	}
}

func switchToDesktop(index int) {
	if currentDesktop != index && index < len(desktops) {
		for _, w := range desktopToWins[currentDesktop] {
			managedWindows[w].Unmap()
		}
		for _, w := range desktopToWins[index] {
			win := managedWindows[w]
			if !win.IsHidden() {
				managedWindows[w].Map()
			}
		}
		currentDesktop = index
		setCurrentDesktop()
	}
}

func SetDesktopNames(names []string) {
	for i, name := range names {
		if i < len(desktops) {
			desktops[i] = name
		}
	}
	if len(names) > len(desktops) {
		setDesktopNames(names)
	} else {
		setDesktopNames(desktops)
	}
}

func mouseShortcutsChanged() {
	for _, win := range managedWindows {
		win.SetupMouseEvents(moveDragShortcut, resizeDragShortcut)
	}
}
