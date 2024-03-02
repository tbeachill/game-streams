package commands

import "github.com/bwmarrin/discordgo"

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "list-streams",
		Description: "List all upcoming streams",
	},
}
