package commands

import "github.com/bwmarrin/discordgo"

// admin is the permission level for an administrator. This is used to set the
// permissions for the help and settings commands.
var admin int64 = discordgo.PermissionAdministrator

// boolFalse is used to set the default permissions for the settings command. This is
// required because the DefaultMemberPermissions field expects a pointer to a boolean
// value.
var boolFalse bool = false

// commands is a slice of all the commands that the bot can register with Discord. Each
// command has a name and description, and some commands have options and permissions.
var commands = []*discordgo.ApplicationCommand{
	{
		Name:         "streams",
		Description:  "List upcoming streams for all platforms",
		DMPermission: &boolFalse,
	},
	{
		Name:         "streaminfo",
		Description:  "Get more information about a specific stream by name",
		DMPermission: &boolFalse,
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
		DMPermission:             &boolFalse,
		DefaultMemberPermissions: &admin,
	},
	{
		Name:                     "suggest",
		Description:              "Suggest a stream to add to the bots database",
		DMPermission:             &boolFalse,
		DefaultMemberPermissions: &admin,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "name",
				Description: "The name of the stream",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "date",
				Description: "The date of the stream",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "The URL of the stream or information about the stream",
				Required:    true,
			},
		},
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
				Name:        "vr",
				Description: "Enable or disable VR stream announcements",
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
