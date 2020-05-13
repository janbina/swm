package main

import (
	"fmt"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
)

type movement struct {
	command      []string
	xDiff, yDiff int
}

func testMovingCommand() int {
	errorCnt := 0

	win := createWindow()
	sleepMillis(100)
	winId := fmt.Sprintf("%d", win.Id)

	initGeom := geom(win)
	prevGeom := initGeom
	var newGeom xrect.Rect

	movements := []movement{
		// Small amount
		{command: []string{"-w", "1"}, xDiff: -1, yDiff: 0},
		{command: []string{"-e", "1"}, xDiff: 1, yDiff: 0},
		{command: []string{"-n", "1"}, xDiff: 0, yDiff: -1},
		{command: []string{"-s", "1"}, xDiff: 0, yDiff: 1},
		// Normal amount
		{command: []string{"-w", "20"}, xDiff: -20, yDiff: 0},
		{command: []string{"-e", "20"}, xDiff: 20, yDiff: 0},
		{command: []string{"-n", "20"}, xDiff: 0, yDiff: -20},
		{command: []string{"-s", "20"}, xDiff: 0, yDiff: 20},
		// Large amount
		{command: []string{"-w", "200"}, xDiff: -200, yDiff: 0},
		{command: []string{"-e", "200"}, xDiff: 200, yDiff: 0},
		{command: []string{"-n", "200"}, xDiff: 0, yDiff: -200},
		{command: []string{"-s", "200"}, xDiff: 0, yDiff: 200},
		// Negative amount
		{command: []string{"-w", "-20"}, xDiff: 20, yDiff: 0},
		{command: []string{"-e", "-20"}, xDiff: -20, yDiff: 0},
		{command: []string{"-n", "-20"}, xDiff: 0, yDiff: 20},
		{command: []string{"-s", "-20"}, xDiff: 0, yDiff: -20},
		// Combinations
		{command: []string{"-w", "20", "-n", "10"}, xDiff: -20, yDiff: -10},
		{command: []string{"-w", "20", "-s", "10"}, xDiff: -20, yDiff: 10},
		{command: []string{"-e", "20", "-n", "10"}, xDiff: 20, yDiff: -10},
		{command: []string{"-e", "20", "-s", "10"}, xDiff: 20, yDiff: 10},
		{command: []string{"-w", "20", "-e", "10"}, xDiff: -10, yDiff: 0},
		{command: []string{"-n", "20", "-s", "10"}, xDiff: 0, yDiff: -10},
	}

	for _, move := range movements {
		for i := 0; i < 10; i++ {
			swmctl(append([]string{"move", "-id", winId}, move.command...)...)
			sleepMillis(50)
			newGeom = geom(win)
			assertEquals(prevGeom.X() + move.xDiff, newGeom.X(), "invalid x coord", &errorCnt)
			assertEquals(prevGeom.Y() + move.yDiff, newGeom.Y(), "invalid y coord", &errorCnt)
			prevGeom = newGeom
		}
		win.Move(initGeom.X(), initGeom.Y())
		prevGeom = initGeom
	}

	win.Destroy()

	return errorCnt
}

func geom(win *xwindow.Window) xrect.Rect {
	r, e := win.DecorGeometry()
	if e != nil {
		return xrect.New(0, 0, 1, 1)
	}
	return r
}
