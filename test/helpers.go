package main

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
	"log"
	"os/exec"
	"time"
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

	win.Map()

	return win
}

func createWindows(count int) []*xwindow.Window {
	wins := make([]*xwindow.Window, count)
	for i := range wins {
		wins[i] = createWindow()
		sleepMillis(100)
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

func sleepMillis(millis time.Duration) {
	time.Sleep(time.Millisecond * millis)
}

func assert(val bool, msg string, errorCnt *int) {
	if !val {
		_ = errorLogger.Output(2, msg)
		*errorCnt++
	}
}

func assertActive(win *xwindow.Window, errorCnt *int) {
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
