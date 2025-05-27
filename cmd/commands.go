package cmd

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
)

type Command struct {
	help    string
	example string
	minArgs int
	maxArgs int // -1 means unlimited
	f       func([]string) (string, bool)
}

var cmdList = map[string]Command{
	"help": {
		help: "Show help information about a command",
		example: "help join",
		minArgs: 1,
		maxArgs: 1,
	},
	"echo": {
		help: "Echo back arguments",
		example: "echo hello world",
		minArgs: 0,
		maxArgs: -1,
		f: cmdEcho,
	},
}

func handleCommand(cmd string) (string, bool) {
	args := strings.Split(cmd, " ")
	c := cmdList[args[0]]

	if c.help == "" {
		return args[0] + ": Unknown command", true
	}

	if len(args) < c.minArgs+1 {
		return args[0] + ": Requires more arguments\nSee /help " + args[0], true
	}

	if c.maxArgs >= 0 && len(args) > c.maxArgs+1 {
		return args[0] + ": Too many arguments\nSee /help " + args[0], true
	}

	// Can't put it inside cmdList because of self-reference
	if args[0] != "help" {
		return c.f(args)
	}

	c = cmdList[args[1]]
	if c.help == "" {
		return args[1] + ": No such command", true
	}
	return args[1] + ": " + c.help + "\nExample: /" + c.example + "\n", false
}

func cmdEcho(args []string) (string, bool) {
	return strings.Join(args[1:], " "), false
}
