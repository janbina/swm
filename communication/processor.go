package communication

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/janbina/swm/windowmanager"
	"strconv"
	"strings"
)

type commandFunc func([]string) string

var commands = map[string]commandFunc{
	"shutdown":   shutdownCommand,
	"destroywin": destroyWindowCommand,
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
