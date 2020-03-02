package communication

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/janbina/swm/windowmanager"
	"strconv"
	"strings"
)

func processCommand(msg string) string {
	args := strings.Fields(msg)

	if len(args) == 0 {
		return "No command"
	}

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "shutdown":
		shutdownCommand()
	case "destroywin":
		return destroyWindowCommand(commandArgs)
	default:
		return fmt.Sprintf("Unknown command: %s", command)
	}

	return ""
}

func shutdownCommand() {
	windowmanager.Shutdown()
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
