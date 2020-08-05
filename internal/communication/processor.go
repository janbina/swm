package communication

import (
	"fmt"
	"sort"
	"strings"

	"github.com/janbina/swm/internal/log"
	"github.com/mattn/go-shellwords"
)

type Action func([]string) string

var commands = make(map[string]Action)

func RegisterCommand(command string, action Action) error {
	return RegisterCommands(map[string]Action{command: action})
}

func RegisterCommands(new map[string]Action) error {
	var used []string

	for command, action := range new {
		if _, ok := commands[command]; ok {
			used = append(used, command)
			continue
		}
		commands[command] = action
	}

	if len(used) > 0 {
		return fmt.Errorf("commands %s are already registered", used)
	}
	return nil
}

func processCommand(msg string) string {
	log.Infof("Got command from swmctl: %s", msg)

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
