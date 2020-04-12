package main

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
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

func swmctl(command string) {
	_, e := swmctlOut(command)
	if e != nil {
		log.Fatalf("Error running swmctl command %s: %s", command, e)
	}
}

func swmctlOut(command string) (string, error) {
	cmd := exec.Command("./swmctl", command)
	out, err := cmd.Output()
	return string(out), err
}

func sleepMillis(millis time.Duration) {
	time.Sleep(time.Millisecond * millis)
}

func assert(val bool, msg string, errorCnt *int) {
	if !val {
		_ = errorLogger.Output(3, msg)
		*errorCnt++
	}
}
