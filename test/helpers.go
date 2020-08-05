package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
)

func createWindow() *xwindow.Window {
	win, err := xwindow.Generate(X)
	if err != nil {
		log.Fatal(err)
	}

	win.Create(X.RootWin(), 0, 0, 200, 200, xproto.CwBackPixel, uint32(0xff0000))

	win.WMGracefulClose(func(w *xwindow.Window) {
		xevent.Detach(w.X, w.Id)
		w.Destroy()
	})

	_ = win.Listen(
		xproto.EventMaskFocusChange,
		xproto.EventMaskStructureNotify,
		xproto.EventMaskPropertyChange,
		xproto.EventMaskFocusChange,
	)

	win.Map()

	active, reparented, mapped := false, false, false
	waitForEvent(func(event xgb.Event) bool {
		switch e := event.(type) {
		case xproto.ReparentNotifyEvent:
			if e.Event == win.Id {
				reparented = true
			}
		case xproto.MapNotifyEvent:
			if e.Event == win.Id {
				mapped = true
			}
		case xproto.PropertyNotifyEvent:
			atom, _ := xprop.Atm(X, "_NET_ACTIVE_WINDOW")
			if e.Window == X.RootWin() && atom == e.Atom && getActiveWindow() == win.Id {
				active = true
			}
		}
		return reparented && mapped && active
	})

	return win
}

func createWindows(count int) []*xwindow.Window {
	wins := make([]*xwindow.Window, count)
	for i := range wins {
		wins[i] = createWindow()
	}
	return wins
}

func destroyWindows(wins []*xwindow.Window) {
	for _, win := range wins {
		win.Destroy()
	}
}

func getActiveWindow() xproto.Window {
	w, _ := ewmh.ActiveWindowGet(X)
	return w
}

func geom(win *xwindow.Window) xrect.Rect {
	r, e := win.DecorGeometry()
	if e != nil {
		return xrect.New(0, 0, 1, 1)
	}
	return r
}

func intStr(i int) string {
	return fmt.Sprintf("%d", i)
}

func floatStr(f float64) string {
	return fmt.Sprintf("%f", f)
}

func swmctl(args ...string) {
	_, e := swmctlOut(args...)
	if e != nil {
		log.Fatalf("Error running swmctl command %s: %s", args, e)
	}
}

func swmctlOut(args ...string) (string, error) {
	cmd := exec.Command("./swmctl", args...)
	out, err := cmd.Output()
	return string(out), err
}

func assert(val bool, msg string, errorCnt *int) {
	if !val {
		_ = errorLogger.Output(2, msg)
		*errorCnt++
	}
}

func assertActive(win *xwindow.Window, errorCnt *int) {
	if win.Id == getActiveWindow() {
		return
	}
	waitForActive(win.Id)
	if win.Id != getActiveWindow() {
		_ = errorLogger.Output(2, "Incorrect active window")
		*errorCnt++
	}
}

func assertEquals(expected, actual int, msg string, errorCnt *int) {
	if expected != actual {
		_ = errorLogger.Output(2, fmt.Sprintf("%s - expected %d, got %d", msg, expected, actual))
		*errorCnt++
	}
}

func assertSliceEquals(expected, actual []int, msg string, errorCnt *int) {
	if !sliceEquals(expected, actual) {
		_ = errorLogger.Output(2, fmt.Sprintf("%s - expected %d, got %d", msg, expected, actual))
		*errorCnt++
	}
}

func assertGeomEquals(expected, actual xrect.Rect, msg string, errorCnt *int) {
	if expected.X() != actual.X() ||
		expected.Y() != actual.Y() ||
		expected.Width() != actual.Width() ||
		expected.Height() != actual.Height() {
		_ = errorLogger.Output(2, fmt.Sprintf("%s - expected %s, got %s", msg, expected, actual))
		*errorCnt++
	}
}

func addToRect(rect xrect.Rect, xD, yD, wD, hD int) xrect.Rect {
	return xrect.New(rect.X()+xD, rect.Y()+yD, rect.Width()+wD, rect.Height()+hD)
}

func sliceEquals(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// waits for event satisfying et function
func waitForEvent(et func(event xgb.Event) bool) {
	// if no matching event comes in this time, just returns
	timeout := time.After(1 * time.Second)
	for {
		select {
		case <-timeout:
			return
		default:
			for {
				event, err := X.Conn().PollForEvent()
				if et(event) {
					return
				}
				if event == nil && err == nil {
					break
				}
			}
		}
	}
}

func flushEvents() {
	for {
		ev, err := X.Conn().PollForEvent()
		if ev == nil && err == nil {
			break
		}
	}
}

func waitForPropertyChange(id xproto.Window, atomName string) {
	waitForEvent(func(event xgb.Event) bool {
		e, ok := event.(xproto.PropertyNotifyEvent)
		if ok {
			atom, _ := xprop.Atm(X, atomName)
			return (id == 0 || e.Window == id) && (atomName == "" || atom == e.Atom)
		}
		return false
	})
}

func waitForConfigureNotify() {
	waitForEvent(func(event xgb.Event) bool {
		_, ok := event.(xproto.ConfigureNotifyEvent)
		return ok
	})
}

func waitForMapNotify() {
	waitForEvent(func(event xgb.Event) bool {
		_, ok := event.(xproto.MapNotifyEvent)
		return ok
	})
}

func waitForUnmapNotify() {
	waitForEvent(func(event xgb.Event) bool {
		_, ok := event.(xproto.UnmapNotifyEvent)
		return ok
	})
}

func waitForActive(id xproto.Window) {
	waitForEvent(func(event xgb.Event) bool {
		e, ok := event.(xproto.PropertyNotifyEvent)
		if ok {
			atom, _ := xprop.Atm(X, "_NET_ACTIVE_WINDOW")
			return e.Window == X.RootWin() && atom == e.Atom && (id == 0 || getActiveWindow() == id)
		}
		return false
	})
}

func repeat(times int, action func()) {
	for i := 0; i < times; i++ {
		action()
	}
}

func isWinIconified(win *xwindow.Window) bool {
	state, _ := icccm.WmStateGet(X, win.Id)
	return state.State == icccm.StateIconic
}

func isWinMapped(win *xwindow.Window) bool {
	state, _ := icccm.WmStateGet(X, win.Id)
	return state.State == icccm.StateNormal
}
