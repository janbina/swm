package main

import (
	"github.com/BurntSushi/xgbutil/xrect"
)

func testResizingCommand() int {
	errorCnt := 0

	win := createWindow()
	winId := intStr(int(win.Id))

	initWinGeom, _ := win.Geometry()
	initGeom := geom(win)
	prevGeom := initGeom
	var newGeom xrect.Rect

	movements := []moveresize{
		// Small amount
		{command: []string{"-w", "1"}, xD: -1, yD: 0, wD: 1, hD: 0},
		{command: []string{"-e", "1"}, xD: 0, yD: 0, wD: 1, hD: 0},
		{command: []string{"-n", "1"}, xD: 0, yD: -1, wD: 0, hD: 1},
		{command: []string{"-s", "1"}, xD: 0, yD: 0, wD: 0, hD: 1},
		//// Normal amount
		{command: []string{"-w", "20"}, xD: -20, yD: 0, wD: 20, hD: 0},
		{command: []string{"-e", "20"}, xD: 0, yD: 0, wD: 20, hD: 0},
		{command: []string{"-n", "20"}, xD: 0, yD: -20, wD: 0, hD: 20},
		{command: []string{"-s", "20"}, xD: 0, yD: 0, wD: 0, hD: 20},
		//// Large amount
		{command: []string{"-w", "200"}, xD: -200, yD: 0, wD: 200, hD: 0},
		{command: []string{"-e", "200"}, xD: 0, yD: 0, wD: 200, hD: 0},
		{command: []string{"-n", "200"}, xD: 0, yD: -200, wD: 0, hD: 200},
		{command: []string{"-s", "200"}, xD: 0, yD: 0, wD: 0, hD: 200},
		//// Negative amount
		{command: []string{"-w", "-10"}, xD: 10, yD: 0, wD: -10, hD: 0},
		{command: []string{"-e", "-10"}, xD: 0, yD: 0, wD: -10, hD: 0},
		{command: []string{"-n", "-10"}, xD: 0, yD: 10, wD: 0, hD: -10},
		{command: []string{"-s", "-10"}, xD: 0, yD: 0, wD: 0, hD: -10},
		//// Combinations
		{command: []string{"-w", "20", "-n", "10"}, xD: -20, yD: -10, wD: 20, hD: 10},
		{command: []string{"-w", "20", "-s", "10"}, xD: -20, yD: 0, wD: 20, hD: 10},
		{command: []string{"-e", "20", "-n", "10"}, xD: 0, yD: -10, wD: 20, hD: 10},
		{command: []string{"-e", "20", "-s", "10"}, xD: 0, yD: 0, wD: 20, hD: 10},
		{command: []string{"-w", "20", "-e", "10"}, xD: -20, yD: 0, wD: 30, hD: 0},
		{command: []string{"-n", "20", "-s", "10"}, xD: 0, yD: -20, wD: 0, hD: 30},
		{command: []string{"-w", "20", "-e", "15", "-n", "10", "-s", "5"}, xD: -20, yD: -10, wD: 35, hD: 15},
		{command: []string{"-w", "-5", "-e", "-5", "-n", "-5", "-s", "-5"}, xD: 5, yD: 5, wD: -10, hD: -10},
	}

	for _, move := range movements {
		for i := 0; i < 10; i++ {
			swmctl(append([]string{"resize", "-id", winId}, move.command...)...)
			waitForConfigureNotify()
			newGeom = geom(win)
			assertGeomEquals(
				addToRect(prevGeom, move.xD, move.yD, move.wD, move.hD),
				newGeom, "invalid geometry", &errorCnt,
			)
			prevGeom = newGeom
		}
		win.MoveResize(initGeom.X(), initGeom.Y(), initWinGeom.Width(), initWinGeom.Height())
		prevGeom = initGeom
	}

	win.Destroy()

	return errorCnt
}
