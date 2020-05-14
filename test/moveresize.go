package main

import (
	"fmt"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
	"math"
)

func testMoveResizeCommand() int {
	errorCnt := 0

	win := createWindow()
	winId := intStr(int(win.Id))

	g, _ := xwindow.New(X, X.RootWin()).Geometry()
	screenGeom := screenGeom{rect: g}
	winGeom := geom(win)
	var newGeom xrect.Rect

	width03 := int(float64(screenGeom.Width()) * .3)
	height03 := int(float64(screenGeom.Height()) * .3)
	movements := []moveresize{
		// default values, top left corner, same size
		{
			command: []string{},
			xD:      screenGeom.X(), yD: screenGeom.Y(), wD: winGeom.Width(), hD: winGeom.Height(),
		},
		// different origin
		{
			command: []string{"-o", "nw", "-x", "10", "-y", "10"},
			xD:      screenGeom.X() + 10, yD: screenGeom.Y() + 10, wD: winGeom.Width(), hD: winGeom.Height(),
		},
		{
			command: []string{"-o", "ne", "-x", "10", "-y", "10"},
			xD:      screenGeom.EndX() - winGeom.Width() - 10, yD: screenGeom.Y() + 10, wD: winGeom.Width(), hD: winGeom.Height(),
		},
		{
			command: []string{"-o", "sw", "-x", "10", "-y", "10"},
			xD:      screenGeom.X() + 10, yD: screenGeom.EndY() - winGeom.Height() - 10, wD: winGeom.Width(), hD: winGeom.Height(),
		},
		{
			command: []string{"-o", "se", "-x", "10", "-y", "10"},
			xD:      screenGeom.EndX() - winGeom.Width() - 10, yD: screenGeom.EndY() - winGeom.Height() - 10, wD: winGeom.Width(), hD: winGeom.Height(),
		},
		{
			command: []string{"-o", "c", "-x", "10", "-y", "10"},
			xD:      screenGeom.EndX()/2 - winGeom.Width()/2 + 10, yD: screenGeom.EndY()/2 - winGeom.Height()/2 + 10, wD: winGeom.Width(), hD: winGeom.Height(),
		},
		// leaving out one direction in origin defaults to center
		{
			command: []string{"-o", "n", "-x", "10", "-y", "10"},
			xD:      screenGeom.EndX()/2 - winGeom.Width()/2 + 10, yD: screenGeom.Y() + 10, wD: winGeom.Width(), hD: winGeom.Height(),
		},
		{
			command: []string{"-o", "s", "-x", "10", "-y", "10"},
			xD:      screenGeom.EndX()/2 - winGeom.Width()/2 + 10, yD: screenGeom.EndY() - winGeom.Height() - 10, wD: winGeom.Width(), hD: winGeom.Height(),
		},
		{
			command: []string{"-o", "w", "-x", "10", "-y", "10"},
			xD:      screenGeom.X() + 10, yD: screenGeom.EndY()/2 - winGeom.Height()/2 + 10, wD: winGeom.Width(), hD: winGeom.Height(),
		},
		{
			command: []string{"-o", "e", "-x", "10", "-y", "10"},
			xD:      screenGeom.EndX() - winGeom.Width() - 10, yD: screenGeom.EndY()/2 - winGeom.Height()/2 + 10, wD: winGeom.Width(), hD: winGeom.Height(),
		},
		// relative coordinates
		{
			command: []string{"-o", "nw", "-xr", ".3", "-yr", ".3"},
			xD:      screenGeom.X() + width03, yD: screenGeom.Y() + height03, wD: winGeom.Width(), hD: winGeom.Height(),
		},
		{
			command: []string{"-o", "ne", "-xr", ".3", "-yr", ".3"},
			xD:      screenGeom.EndX() - winGeom.Width() - width03, yD: screenGeom.Y() + height03, wD: winGeom.Width(), hD: winGeom.Height(),
		},
		{
			command: []string{"-o", "sw", "-xr", ".3", "-yr", ".3"},
			xD:      screenGeom.X() + width03, yD: screenGeom.EndY() - winGeom.Height() - height03, wD: winGeom.Width(), hD: winGeom.Height(),
		},
		{
			command: []string{"-o", "se", "-xr", ".3", "-yr", ".3"},
			xD:      screenGeom.EndX() - winGeom.Width() - width03, yD: screenGeom.EndY() - winGeom.Height() - height03, wD: winGeom.Width(), hD: winGeom.Height(),
		},
		{
			command: []string{"-o", "c", "-xr", ".3", "-yr", ".3"},
			xD:      screenGeom.EndX()/2 - winGeom.Width()/2 + width03, yD: screenGeom.EndY()/2 - winGeom.Height()/2 + height03, wD: winGeom.Width(), hD: winGeom.Height(),
		},
		// combination
		{
			command: []string{"-o", "nw", "-x", "10", "-yr", ".3"},
			xD:      screenGeom.X() + 10, yD: screenGeom.Y() + height03, wD: winGeom.Width(), hD: winGeom.Height(),
		},
	}

	// width and height
	for i := 0; i < 10; i++ {
		w := i*150 + 1
		h := i*140 + 1
		wr := math.Round((float64(i)*.15+.1)*100) / 100
		hr := math.Round((float64(i)*.14+.1)*100) / 100
		movements = append(movements,
			moveresize{
				command: []string{"-w", intStr(w), "-h", intStr(h)},
				xD:      0,
				yD:      0,
				wD:      w,
				hD:      h,
			},
			moveresize{
				command: []string{"-wr", floatStr(wr), "-hr", floatStr(hr)},
				xD:      0,
				yD:      0,
				wD:      int(wr * float64(screenGeom.Width())),
				hD:      int(hr * float64(screenGeom.Height())),
			},
		)
	}

	// position and size combination
	movements = append(movements, []moveresize{
		{
			command: []string{"-o", "nw", "-x", "10", "-y", "10", "-w", "300", "-h", "400"},
			xD:      screenGeom.X() + 10, yD: screenGeom.Y() + 10, wD: 300, hD: 400,
		},
		{
			command: []string{"-o", "nw", "-x", "0", "-y", "0", "-wr", ".5", "-hr", ".5"},
			xD:      screenGeom.X(), yD: screenGeom.Y(), wD: screenGeom.Width()/2, hD: screenGeom.Height()/2,
		},
	}...)

	for i, move := range movements {
		swmctl(append([]string{"moveresize", "-id", winId}, move.command...)...)
		waitForConfigureNotify()
		newGeom = geom(win)
		assertGeomEquals(
			xrect.New(move.xD, move.yD, move.wD, move.hD),
			newGeom,
			fmt.Sprintf("invalid geometry (%d)", i),
			&errorCnt,
		)
	}

	win.Destroy()

	return errorCnt
}

type screenGeom struct {
	rect xrect.Rect
}

func (g screenGeom) X() int {
	return g.rect.X()
}

func (g screenGeom) Y() int {
	return g.rect.Y()
}

func (g screenGeom) Width() int {
	return g.rect.Width()
}

func (g screenGeom) Height() int {
	return g.rect.Height()
}

func (g screenGeom) EndX() int {
	return g.rect.X() + g.rect.Width()
}

func (g screenGeom) EndY() int {
	return g.rect.Y() + g.rect.Height()
}
