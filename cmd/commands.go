package cmd

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type Command struct {
	help    string
	example string
	minArgs int
	maxArgs int // -1 means unlimited
	f       func([]string) string
}

var cmdList = map[string]Command{
	"help": {
		help: "show help information about a command",
		example: "help join",
		minArgs: 1,
		maxArgs: 1,
	},
	"cmdlist": {
		help: "list avaliable commands",
		example: "cmdlist",
		minArgs: 0,
		maxArgs: 0,
	},
	"echo": {
		help: "echo back arguments",
		example: "echo hello world",
		minArgs: 0,
		maxArgs: -1,
		f: cmdEcho,
	},
	"join": {
		help: "join an invite link\nNote: DO NOT USE, WILL FLAG YOUR ACCOUTNS",
		example: "join https://discord.com/invite/VzF9UFn2aB",
		minArgs: 1,
		maxArgs: 1,
		f: cmdJoin,
	},
	"mention": {
		help: "mention a user",
		example: "mention someguy123",
		minArgs: 1,
		maxArgs: 1,
		f: cmdMention,
	},
	"status": {
		help: "show/set current user status.\n        To permanently set user status, use the config file",
		example: "status idle",
		minArgs: 0,
		maxArgs: 1,
		f: cmdStatus,
	},
}

func handleCommand(cmd string) string {
	args := strings.Split(cmd, " ")
	c := cmdList[args[0]]

	if c.help == "" {
		warn("%s: Unknown command", args[0])
		return ""
	}

	if len(args) < c.minArgs+1 {
		warn("%s: Requires more arguments\nSee /help %[0]s", args[0])
		return ""
	}

	if c.maxArgs >= 0 && len(args) > c.maxArgs+1 {
		warn("%s: Too many arguments\nSee /help %[0]s", args[0])
		return ""
	}

	// Can't put them inside cmdList because of self-reference
	if args[0] != "help" && args[0] != "cmdlist" {
		return c.f(args[1:])
	}

	if args[0] == "cmdlist" {
		for x, c := range cmdList {
			inform("/%6s - ", x, c.help)
		}
		return ""
	}

	c = cmdList[args[1]]
	if c.help == "" {
		warn("%s: No such command", args[1])
		return ""
	}
	inform("%s: %s\nExample: /%s", args[1], c.help, c.example)
	return ""
}

func warn(f string, a ...any) {
	app.messagesText.displayInternalMsg(true, f, a...)
}

func inform(f string, a ...any) {
	app.messagesText.displayInternalMsg(false, f, a...)
}

func cmdEcho(args []string) string {
	inform("%s", strings.Join(args, " "))
	return ""
}

func cmdJoin(args []string) string {
	code := strings.TrimPrefix(args[0], "http")
	code = strings.TrimPrefix(code, "s")
	code = strings.TrimPrefix(code, "://")
	code = strings.TrimPrefix(code, "discord.com/")
	code = strings.TrimPrefix(code, "invite/")

	for _, c := range code {
		if !((c >= 'a' && c <= 'z') ||
		     (c >= 'A' && c <= 'Z') ||
		     (c >= '0' && c <= '9')) {
			warn("%s: Invalid invite", args[0])
			return ""
		}
	}

	client := api.NewClient(discordState.Token)
	client.UserAgent = app.cfg.Identify.UserAgent
	_, err := client.JoinInvite(code)
	if err != nil {
		warn("Failed to join: %s: %s", args[0], err.Error())
		return ""
	}
	inform("Successfully joined: %s", args[0])
	return ""
}

func cmdMention(args []string) string {
	var res discord.Member
	guildID := app.guildsTree.selectedGuildID
	q := strings.Join(args, " ")

	// Load user into cabinet
	updates := make(chan *gateway.GuildMembersChunkEvent, 1)
	cancel := discordState.AddHandler(updates)
	discordState.MemberState.SearchMember(guildID, q)
	<-updates
	cancel()

	// Search usernames first
	discordState.MemberStore.Each(guildID, func (m *discord.Member) bool {
		if q == m.User.Username {
			res = *m
			return true
		}
		return false
	})

	if res.User.Username == "" {
		discordState.MemberStore.Each(guildID,
			func (m *discord.Member) bool {
				if q == m.User.DisplayName {
					res = *m
					return true
				}
				return false
			},
		)
	}

	if res.User.Username == "" {
		warn("%s: User not found", q)
		return ""
	}

	inform("%s: %s", res.User.DisplayOrUsername(), res.User.ID.String())
	return "<@" + res.User.ID.String() + ">"
}

func cmdStatus(args []string) string {
	if len(args) == 1 {
		inform("Current status: %s", string(discordState.Status()))
		return ""
	}

	err := discordState.SetStatus(discord.Status(args[0]), nil)
	if err != nil {
		warn("Failed to set status to: %s: ", args[0], err.Error())
		return ""
	}

	inform("Changed current (non-permanent) status to: %s", args[0])
	return ""
}
