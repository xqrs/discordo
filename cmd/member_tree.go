package cmd


import (
	"fmt"
	"log/slog"
	"sort"

	"github.com/ayn2op/discordo/internal/config"
	"github.com/ayn2op/discordo/internal/ui"
	"github.com/ayn2op/tview"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/gdamore/tcell/v2"
)

type memberTree struct {
	*tview.TreeView
	cfg *config.Config
}

func newMemberTree(cfg *config.Config) *memberTree {
	mt := &memberTree{
		TreeView: tview.NewTreeView(),
		cfg:      cfg,
	}

	mt.Box = ui.ConfigureBox(mt.Box, &cfg.Theme)
	mt.SetRoot(tview.NewTreeNode("")).
		SetTopLevel(1).
		SetTitle("Members")
	mt.SetGraphics(false).
		SetPrefixes([]string{""}).
		SetAlign(true)

	return mt
}

func getStatus(guildID discord.GuildID, userID discord.UserID) (string, bool) {
	presence, err := discordState.Cabinet.Presence(guildID, userID)
	if err != nil {
		// Missing presence = offline
		return "[gray] ", true
	}

	switch presence.Status {
	case discord.OnlineStatus:
		return "[green] ", false
	case discord.IdleStatus:
		return "[yellow] ", false
	case discord.DoNotDisturbStatus:
		return "[red] ", false
	case discord.OfflineStatus:
		return "[gray] ", true
	default:
		return "[gray] ", true
	}
}

func (mt *memberTree) Update(guildID discord.GuildID, channelID discord.ChannelID, members []discord.Member) {
	root := mt.GetRoot()
	root.ClearChildren()

	roles, err := discordState.Cabinet.Roles(guildID)
	if err != nil {
		slog.Error("failed to get guild roles", "err", err)
		return
	}

	sort.Slice(roles, func(i, j int) bool {
		if roles[i].Position != roles[j].Position {
			return roles[i].Position > roles[j].Position
		}
		return roles[i].ID < roles[j].ID
	})

	classified := make(map[discord.UserID]bool)

	// ONLINE by roles
	var roleOnline []*tview.TreeNode
	for _, role := range roles {
		if role.ID == discord.RoleID(guildID) || !role.Hoist {
			continue
		}

		for _, m := range members {
			if classified[m.User.ID] {
				continue
			}

			prefix, offline := getStatus(guildID, m.User.ID)
			if offline {
				continue
			}

			if memberHasRole(m, role.ID) && channelHasUser(channelID, m.User.ID){
				name := m.User.DisplayOrUsername()
				memberNode := tview.NewTreeNode(prefix + name).
					SetColor(tcell.GetColor(role.Color.String()))
				roleOnline = append(roleOnline, memberNode)
				classified[m.User.ID] = true
			}
		}

		if len(roleOnline) > 0 {
			roleNode := tview.NewTreeNode(fmt.Sprintf("%s (%d)", role.Name, len(roleOnline))).
				SetColor(tcell.GetColor(role.Color.String()))
			for _, node := range roleOnline {
				roleNode.AddChild(node)
			}
			root.AddChild(roleNode)
		}
	}

	// ONLINE without roles
	for _, m := range members {
		if classified[m.User.ID] {
			continue
		}

		prefix, offline := getStatus(guildID, m.User.ID)
		if offline {
			continue
		}

		name := m.User.DisplayOrUsername()
		root.AddChild(tview.NewTreeNode(prefix + name))
		classified[m.User.ID] = true
	}

	// OFFLINE (everyone else)
	offlineNode := tview.NewTreeNode("Offline")
	root.AddChild(offlineNode)
	for _, m := range members {
		if classified[m.User.ID] {
			continue
		}

		prefix, offline := getStatus(guildID, m.User.ID)
		if !offline {
			continue
		}

		name := m.User.DisplayOrUsername()
		offlineNode.AddChild(tview.NewTreeNode(prefix + name))
		classified[m.User.ID] = true
	}

	root.ExpandAll()
}

// helper
func memberHasRole(member discord.Member, roleID discord.RoleID) bool {
	for _, id := range member.RoleIDs {
		if id == roleID {
			return true
		}
	}
	return false
}
