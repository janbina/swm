package communication

import (
	"flag"
	"fmt"
	"github.com/janbina/swm/desktopmanager"
	"github.com/janbina/swm/window"
	"github.com/janbina/swm/windowmanager"
	"github.com/mattn/go-shellwords"
	"log"
	"strings"
)

var commands = map[string]func([]string) string{
	"shutdown":             shutdownCommand,
	"resize":               resizeCommand,
	"move-drag-shortcut":   moveDragShortcutCommand,
	"resize-drag-shortcut": resizeDragShortcutCommand,
	"moveresize":           moveResizeCommand,
	"move":                 moveCommand,
	"set-desktop-names":    setDesktopNamesCommand,
	"cycle-win":            cycleWinCommand,
	"cycle-win-rev":        cycleWinRevCommand,
	"cycle-win-end":        cycleWinEndCommand,
	"begin-mouse-move":     mouseMoveCommand,
	"begin-mouse-resize":   mouseResizeCommand,
}

func processCommand(msg string) string {
	log.Printf("Got command from swmctl: %s", msg)

	args, _ := shellwords.Parse(msg)

	if len(args) == 0 {
		return "No command"
	}

	command := args[0]
	commandArgs := args[1:]

	if c, ok := commands[command]; !ok {
		return fmt.Sprintf("Unknown command: %s", command)
	} else {
		return c(commandArgs)
	}
}

func shutdownCommand(_ []string) string {
	windowmanager.Shutdown()
	return ""
}

func resizeCommand(args []string) string {
	d := window.Directions{}
	f := flag.NewFlagSet("moveresize", flag.ContinueOnError)
	f.IntVar(&d.Left, "l", 0, "")
	f.IntVar(&d.Bottom, "b", 0, "")
	f.IntVar(&d.Top, "t", 0, "")
	f.IntVar(&d.Right, "r", 0, "")

	if err := f.Parse(args); err != nil {
		return fmt.Sprintf("Error parsing arguments: %s", err)
	}

	windowmanager.ResizeActiveWindow(d)
	return ""
}

func moveDragShortcutCommand(args []string) string {
	if len(args) == 0 {
		return "No shortcut provided"
	}
	s := args[0]
	err := windowmanager.SetMoveDragShortcut(s)
	if err != nil {
		return "Invalid shortcut"
	}
	return ""
}

func resizeDragShortcutCommand(args []string) string {
	if len(args) == 0 {
		return "No shortcut provided"
	}
	s := args[0]
	err := windowmanager.SetResizeDragShortcut(s)
	if err != nil {
		return "Invalid shortcut"
	}
	return ""
}

func moveResizeCommand(args []string) string {
	f := flag.NewFlagSet("moveresize", flag.ContinueOnError)
	anchor := f.String("anchor", "tl", "")
	x := f.Int("x", 0, "")
	y := f.Int("y", 0, "")
	w := f.Int("w", 0, "")
	h := f.Int("h", 0, "")
	xr := f.Float64("xr", 0, "")
	yr := f.Float64("yr", 0, "")
	wr := f.Float64("wr", 0, "")
	hr := f.Float64("hr", 0, "")

	if err := f.Parse(args); err != nil {
		return fmt.Sprintf("Error parsing arguments: %s", err)
	}

	screenGeom, err := windowmanager.GetCurrentScreenGeometryStruts()
	if err != nil {
		return fmt.Sprintf("Cannot get window screen geometry: %s", err)
	}
	winGeom, err := windowmanager.GetActiveWindowGeometry()
	if err != nil {
		return fmt.Sprintf("Cannot get active window geometry: %s", err)
	}
	if *x == 0 {
		*x = int(*xr * float64(screenGeom.Width()))
	}

	if *y == 0 {
		*y = int(*yr * float64(screenGeom.Height()))
	}

	if *w == 0 {
		*w = int(*wr * float64(screenGeom.Width()))
	}

	if *h == 0 {
		*h = int(*hr * float64(screenGeom.Height()))
	}

	if *w <= 0 {
		*w = winGeom.Width()
	}

	if *h <= 0 {
		*h = winGeom.Height()
	}

	var realY int
	if strings.Contains(*anchor, "t") {
		realY = screenGeom.Y() + *y
	} else if strings.Contains(*anchor, "b") {
		realY = screenGeom.Y() + screenGeom.Height() - *y - *h
	} else { //center
		realY = screenGeom.Y() + screenGeom.Height()/2 - *h/2 + *y
	}

	var realX int
	if strings.Contains(*anchor, "l") {
		realX = screenGeom.X() + *x
	} else if strings.Contains(*anchor, "r") {
		realX = screenGeom.X() + screenGeom.Width() - *x - *w
	} else { //center
		realX = screenGeom.X() + screenGeom.Width()/2 - *w/2 + *x
	}

	windowmanager.MoveResizeActiveWindow(realX, realY, *w, *h)

	return ""
}

func moveCommand(args []string) string {
	f := flag.NewFlagSet("move", flag.ContinueOnError)
	l := f.Int("l", 0, "")
	b := f.Int("b", 0, "")
	t := f.Int("t", 0, "")
	r := f.Int("r", 0, "")

	if err := f.Parse(args); err != nil {
		return fmt.Sprintf("Error parsing arguments: %s", err)
	}

	winGeom, err := windowmanager.GetActiveWindowGeometry()
	if err != nil {
		return fmt.Sprintf("Cannot get active window geometry: %s", err)
	}

	dx := *r - *l
	dy := *b - *t

	windowmanager.MoveActiveWindow(winGeom.X()+dx, winGeom.Y()+dy)

	return ""
}

func setDesktopNamesCommand(args []string) string {
	desktopmanager.SetDesktopNames(args)
	return ""
}

func cycleWinCommand(_ []string) string {
	windowmanager.CycleWin()
	return ""
}

func cycleWinRevCommand(_ []string) string {
	windowmanager.CycleWinRev()
	return ""
}

func cycleWinEndCommand(_ []string) string {
	windowmanager.CycleWinEnd()
	return ""
}

func mouseMoveCommand(_ []string) string {
	if err := windowmanager.BeginMouseMoveFromPointer(); err != nil {
		return err.Error()
	}
	return ""
}

func mouseResizeCommand(_ []string) string {
	if err := windowmanager.BeginMouseResizeFromPointer(); err != nil {
		return err.Error()
	}
	return ""
}
