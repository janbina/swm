package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xwindow"
)

var X *xgbutil.XUtil

type test struct {
	name string
	fun  func() int
}

var tests = []test{
	{"cycling", testCycling},
	{"desktop names", testDesktopNames},
	{"group basics", testGroupBasics},
	{"group window creation", testGroupWindowCreation},
	{"group window movement", testGroupWindowMovement},
	{"group visibility", testGroupVisibility},
	{"group membership", testGroupMembership},
	{"moving command", testMovingCommand},
	{"resizing command", testResizingCommand},
	{"moveresize command", testMoveResizeCommand},
	{"window states", testWindowStates},
}

var errorLogger = log.New(os.Stdout, "    error: ", log.Lshortfile)

func main() {
	var err error
	xgb.Logger = log.New(ioutil.Discard, "", 0)
	X, err = xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	defer X.Conn().Close()

	_ = xwindow.New(X, X.RootWin()).Listen(
		xproto.EventMaskPropertyChange,
		xproto.EventMaskSubstructureNotify,
	)

	errorCnt := 0
	for _, t := range tests {
		fmt.Printf("Testing %s ... ", t.name)
		start := time.Now()
		errs := t.fun()
		duration := time.Since(start)
		if errs == 0 {
			fmt.Printf("OK, took %s\n", duration)
		} else {
			fmt.Printf("Errors in %s: %d, took %s\n", t.name, errs, duration)
		}
		errorCnt += errs
	}

	if errorCnt > 0 {
		fmt.Printf("Total number of errors: %d\n", errorCnt)
		os.Exit(1)
	} else {
		fmt.Printf("All is good\n")
	}
	os.Exit(1)
}
