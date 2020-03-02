package communication

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/janbina/swm/window"
	"github.com/janbina/swm/windowmanager"
	"strconv"
	"strings"
)

type commandFunc func([]string) string

var commands = map[string]commandFunc{
	"shutdown":   shutdownCommand,
	"destroywin": destroyWindowCommand,
	"resize":     resizeCommand,
}

func processCommand(msg string) string {
	args := strings.Fields(msg)

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

func destroyWindowCommand(args []string) string {
	if len(args) == 0 {
		windowmanager.DestroyActiveWindow()
	} else {
		if win, err := strconv.Atoi(args[0]); err != nil {
			return fmt.Sprintf("Expected window id (int) as first argument, got %s", args[0])
		} else {
			windowmanager.DestroyWindow(xproto.Window(win))
		}
	}
	return ""
}

func resizeCommand(args []string) string {
	d := window.Directions{}
	for i := 0; i < len(args) - 1; i += 2 {
		name := args[i]
		value, err := strconv.Atoi(args[i+1])
		if err != nil {
			return fmt.Sprintf("Invalid value for argument %s. Expected int, got %s", name, args[i+1])
		}
		switch name {
		case "l", "left":
			d.Left = value
		case "r", "right":
			d.Right = value
		case "t", "top":
			d.Top = value
		case "b", "bottom":
			d.Bottom = value
		}
	}
	windowmanager.ResizeActiveWindow(d)
	return ""
}
