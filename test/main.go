package main

import (
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgbutil"
	"io/ioutil"
	"log"
	"os"
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

	errorCnt := 0
	for _, t := range tests {
		fmt.Printf("Testing %s ... ", t.name)
		errs := t.fun()
		if errs == 0 {
			fmt.Printf("OK\n")
		} else {
			fmt.Printf("Errors in %s: %d\n", t.name, errs)
		}
		errorCnt += errs
	}

	if errorCnt > 0 {
		fmt.Printf("Total number of errors: %d\n", errorCnt)
		os.Exit(1)
	} else {
		fmt.Printf("All is good\n")
	}
}
