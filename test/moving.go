package main

import (
	"github.com/BurntSushi/xgbutil/xrect"
)

type moveresize struct {
	command        []string
	xD, yD, wD, hD int
}

func testMovingCommand() int {
	errorCnt := 0

	win := createWindow()
	winId := intStr(int(win.Id))

	initGeom := geom(win)
	prevGeom := initGeom
	var newGeom xrect.Rect

	movements := []moveresize{
		// Small amount
		{command: []string{"-w", "1"}, xD: -1, yD: 0},
		{command: []string{"-e", "1"}, xD: 1, yD: 0},
		{command: []string{"-n", "1"}, xD: 0, yD: -1},
		{command: []string{"-s", "1"}, xD: 0, yD: 1},
		// Normal amount
		{command: []string{"-w", "20"}, xD: -20, yD: 0},
		{command: []string{"-e", "20"}, xD: 20, yD: 0},
		{command: []string{"-n", "20"}, xD: 0, yD: -20},
		{command: []string{"-s", "20"}, xD: 0, yD: 20},
		// Large amount
		{command: []string{"-w", "200"}, xD: -200, yD: 0},
		{command: []string{"-e", "200"}, xD: 200, yD: 0},
		{command: []string{"-n", "200"}, xD: 0, yD: -200},
		{command: []string{"-s", "200"}, xD: 0, yD: 200},
		// Negative amount
		{command: []string{"-w", "-20"}, xD: 20, yD: 0},
		{command: []string{"-e", "-20"}, xD: -20, yD: 0},
		{command: []string{"-n", "-20"}, xD: 0, yD: 20},
		{command: []string{"-s", "-20"}, xD: 0, yD: -20},
		// Combinations
		{command: []string{"-w", "20", "-n", "10"}, xD: -20, yD: -10},
		{command: []string{"-w", "20", "-s", "10"}, xD: -20, yD: 10},
		{command: []string{"-e", "20", "-n", "10"}, xD: 20, yD: -10},
		{command: []string{"-e", "20", "-s", "10"}, xD: 20, yD: 10},
		{command: []string{"-w", "20", "-e", "10"}, xD: -10, yD: 0},
		{command: []string{"-n", "20", "-s", "10"}, xD: 0, yD: -10},
	}

	for _, move := range movements {
		for i := 0; i < 10; i++ {
			swmctl(append([]string{"move", "-id", winId}, move.command...)...)
			waitForConfigureNotify()
			newGeom = geom(win)
			assertGeomEquals(
				addToRect(prevGeom, move.xD, move.yD, move.wD, move.hD),
				newGeom, "invalid geometry", &errorCnt,
			)
			prevGeom = newGeom
		}
		win.Move(initGeom.X(), initGeom.Y())
		prevGeom = initGeom
	}

	win.Destroy()

	return errorCnt
}
