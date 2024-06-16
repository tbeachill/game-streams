package commands

import "github.com/bwmarrin/discordgo"

var admin int64 = discordgo.PermissionAdministrator
var boolFalse bool = false

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "streams",
		Description: "List upcoming streams for all platforms",
	},
	{
		Name:        "streaminfo",
		Description: "Get more information about a specific stream by name",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "name",
				Description: "The name of the stream",
				Required:    true,
			},
		},
	},
	{
		Name:                     "help",
		Description:              "Get help with the bot",
		DefaultMemberPermissions: &admin,
	},
	{
		Name:                     "settings",
		Description:              "Change bot settings",
		DefaultMemberPermissions: &admin,
		DMPermission:             &boolFalse,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Set the channel for announcing when a stream starts",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "Set the role to ping when a stream starts",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "playstation",
				Description: "Enable or disable Playstation stream announcements",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "xbox",
				Description: "Enable or disable Xbox stream announcements",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "nintendo",
				Description: "Enable or disable Nintendo stream announcements",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "pc",
				Description: "Enable or disable PC stream announcements",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "reset",
				Description: "Reset all settings to default",
				Required:    false,
			},
		},
	},
}
