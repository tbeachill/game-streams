package commands

import "github.com/bwmarrin/discordgo"

//TODO: Create a command to select which platforms to follow
// TODO: Create a command to select which channel to post live streams to

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "list-streams",
		Description: "List all upcoming streams",
	},
}
