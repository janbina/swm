package communication

import (
	"flag"
	"fmt"
	"github.com/janbina/swm/config"
	"github.com/janbina/swm/groupmanager"
	"github.com/janbina/swm/windowmanager"
	"github.com/mattn/go-shellwords"
	"log"
	"sort"
	"strconv"
	"strings"
)

var commands = map[string]func([]string) string{
	"shutdown":             shutdownCommand,
	"move":                 moveCommand,
	"resize":               resizeCommand,
	"moveresize":           moveResizeCommand,
	"cycle-win":            cycleWinCommand,
	"cycle-win-rev":        cycleWinRevCommand,
	"cycle-win-end":        cycleWinEndCommand,
	"set-desktop-names":    setDesktopNamesCommand,
	"begin-mouse-move":     mouseMoveCommand,
	"begin-mouse-resize":   mouseResizeCommand,
	"config":               configCommand,
	"group":                groupCommand,
}

func processCommand(msg string) string {
	log.Printf("Got command from swmctl: %s", msg)

	args, _ := shellwords.Parse(msg)

	if len(args) == 0 {
		return printUsage("No command")
	}

	command := args[0]
	commandArgs := args[1:]

	if c, ok := commands[command]; !ok {
		return printUsage(fmt.Sprintf("Unknown command: %s", command))
	} else {
		return c(commandArgs)
	}
}

func printUsage(firstLine string) string {
	var r strings.Builder
	r.WriteString(firstLine)
	r.WriteByte('\n')
	r.WriteString("Usage: swmctl <cmd> <args>\n")
	r.WriteString("Available commands:\n")

	keys := make([]string, 0, len(commands))
	for k := range commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		r.WriteString(fmt.Sprintf("\t%s\n", k))
	}

	return r.String()
}

func shutdownCommand(_ []string) string {
	windowmanager.Shutdown()
	return ""
}

func moveCommand(args []string) string {
	f := flag.NewFlagSet("move", flag.ContinueOnError)
	id := f.Int("id", 0, "")
	west := f.Int("w", 0, "")
	south := f.Int("s", 0, "")
	north := f.Int("n", 0, "")
	east := f.Int("e", 0, "")

	if err := f.Parse(args); err != nil {
		return fmt.Sprintf("Error parsing arguments: %s", err)
	}

	winGeom, err := windowmanager.GetWindowGeometry(*id)
	if err != nil {
		return fmt.Sprintf("Cannot get active window geometry: %s", err)
	}

	dx := *east - *west
	dy := *south - *north

	if err := windowmanager.MoveWindow(*id, winGeom.X()+dx, winGeom.Y()+dy); err != nil {
		return err.Error()
	}
	return ""
}

func resizeCommand(args []string) string {
	f := flag.NewFlagSet("resize", flag.ContinueOnError)
	id := f.Int("id", 0, "")
	west := f.Int("w", 0, "")
	south := f.Int("s", 0, "")
	north := f.Int("n", 0, "")
	east := f.Int("e", 0, "")

	if err := f.Parse(args); err != nil {
		return fmt.Sprintf("Error parsing arguments: %s", err)
	}

	winGeom, err := windowmanager.GetWindowGeometry(*id)
	if err != nil {
		return fmt.Sprintf("Cannot get active window geometry: %s", err)
	}

	x := winGeom.X() - *west
	y := winGeom.Y() - *north

	width := winGeom.Width() + *west + *east
	height := winGeom.Height() + *north + *south

	if err := windowmanager.MoveResizeWindow(*id, x, y, width, height); err != nil {
		return err.Error()
	}
	return ""
}

func moveResizeCommand(args []string) string {
	f := flag.NewFlagSet("moveresize", flag.ContinueOnError)
	id := f.Int("id", 0, "")
	gravity := f.String("g", "nw", "")
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

	screenGeom, err := windowmanager.GetWindowScreenGeometryStruts(*id)
	if err != nil {
		return fmt.Sprintf("Cannot get window screen geometry: %s", err)
	}
	winGeom, err := windowmanager.GetWindowGeometry(*id)
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
	if strings.Contains(*gravity, "n") {
		realY = screenGeom.Y() + *y
	} else if strings.Contains(*gravity, "s") {
		realY = screenGeom.Y() + screenGeom.Height() - *y - *h
	} else { //center
		realY = screenGeom.Y() + screenGeom.Height()/2 - *h/2 + *y
	}

	var realX int
	if strings.Contains(*gravity, "w") {
		realX = screenGeom.X() + *x
	} else if strings.Contains(*gravity, "e") {
		realX = screenGeom.X() + screenGeom.Width() - *x - *w
	} else { //center
		realX = screenGeom.X() + screenGeom.Width()/2 - *w/2 + *x
	}

	if err := windowmanager.MoveResizeWindow(*id, realX, realY, *w, *h); err != nil {
		return err.Error()
	}
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

func setDesktopNamesCommand(args []string) string {
	groupmanager.SetGroupNames(args)
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

func configCommand(args []string) string {
	if len(args) == 0 {
		return "Nothing to configure"
	}
	switch args[0] {
	case "border":
		s, n, ac, att, err := parseBorderConfig(args[1:])
		if err != nil {
			return err.Error()
		}
		config.SetAllBorders(s, n, ac, att)
	case "border-top":
		s, n, ac, att, err := parseBorderConfig(args[1:])
		if err != nil {
			return err.Error()
		}
		config.SetTopBorder(s, n, ac, att)
	case "border-bottom":
		s, n, ac, att, err := parseBorderConfig(args[1:])
		if err != nil {
			return err.Error()
		}
		config.SetBottomBorder(s, n, ac, att)
	case "border-left":
		s, n, ac, att, err := parseBorderConfig(args[1:])
		if err != nil {
			return err.Error()
		}
		config.SetLeftBorder(s, n, ac, att)
	case "border-right":
		s, n, ac, att, err := parseBorderConfig(args[1:])
		if err != nil {
			return err.Error()
		}
		config.SetRightBorder(s, n, ac, att)
	case "move-drag-shortcut":
		if len(args) < 2 {
			return "No shortcut provided"
		}
		s := args[1]
		err := windowmanager.SetMoveDragShortcut(s)
		if err != nil {
			return "Invalid shortcut"
		}
	case "resize-drag-shortcut":
		if len(args) < 2 {
			return "No shortcut provided"
		}
		s := args[1]
		err := windowmanager.SetResizeDragShortcut(s)
		if err != nil {
			return "Invalid shortcut"
		}
	default:
		return "Unsupported config argument"
	}
	return ""
}

func groupCommand(args []string) string {
	if len(args) == 0 {
		return "No arguments for group command"
	}
	switch args[0] {
	case "mode":
		if len(args) < 2 {
			return "No group mode specified"
		}
		switch args[1] {
		case "sticky":
			groupmanager.GroupMode = groupmanager.ModeSticky
		case "auto":
			groupmanager.GroupMode = groupmanager.ModeAuto
		default:
			return "Unsupported group mode"
		}
	case "toggle", "show", "hide", "only", "set":
		if len(args) < 2 {
			return "No group id to work with"
		}
		if id, err := strconv.Atoi(args[1]); err != nil {
			return "Invalid group id"
		} else {
			switch args[0] {
			case "toggle":
				windowmanager.ToggleGroupVisibility(id)
			case "show":
				windowmanager.ShowGroup(id)
			case "hide":
				windowmanager.HideGroup(id)
			case "only":
				windowmanager.ShowGroupOnly(id)
			case "set":
				if err := windowmanager.SetGroupForActiveWindow(id); err != nil {
					return err.Error()
				}
			default:
				panic("Unreachable")
			}
		}
	case "get-visible":
		var r strings.Builder
		for i, id := range groupmanager.GetVisibleGroups() {
			if i > 0 {
				r.WriteByte('\n')
			}
			r.WriteString(fmt.Sprintf("%d", id))
		}
		return r.String()
	case "get":
		g, err := windowmanager.GetActiveWindowGroups()
		if err != nil {
			return err.Error()
		}
		var r strings.Builder
		for i, id := range g {
			if i > 0 {
				r.WriteByte('\n')
			}
			r.WriteString(fmt.Sprintf("%d", id))
		}
		return r.String()
	default:
		return "Unsupported group argument"
	}
	return ""
}

func parseBorderConfig(args []string) (int, uint32, uint32, uint32, error) {
	if len(args) < 4 {
		return 0, 0, 0, 0, fmt.Errorf("too few arguments for border config")
	}
	var err error
	s, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, 0, 0, 0, err
	}
	n, err := hex2int(args[1])
	if err != nil {
		return 0, 0, 0, 0, err
	}
	ac, err := hex2int(args[2])
	if err != nil {
		return 0, 0, 0, 0, err
	}
	att, err := hex2int(args[3])
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return s, uint32(n), uint32(ac), uint32(att), nil
}

func hex2int(hexStr string) (uint64, error) {
	hexStr = strings.Replace(hexStr, "0x", "", 1)
	hexStr = strings.Replace(hexStr, "#", "", 1)

	return strconv.ParseUint(hexStr, 16, 64)
}
