package commands

import (
	"strconv"

	"github.com/bwmarrin/discordgo"

	"gamestreambot/db"
	"gamestreambot/streams"
	"gamestreambot/utils"
)

// map of command names to their respective functions
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"streams":  listStreams,
	"help":     help,
	"settings": settings,
}

// list all upcoming streams
func listStreams(s *discordgo.Session, i *discordgo.InteractionCreate) {
	content, getErr := streams.StreamList()
	if getErr != nil {
		utils.EWLogger.WithPrefix(" CMND").Error("error getting stream list", "err", getErr)
		content = "An error occurred."
	}

	respondErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if respondErr != nil {
		utils.EWLogger.WithPrefix(" CMND").Error("error responding to interaction", "cmd", i.ApplicationCommandData().Name, "err", respondErr)
	}
}

// get help with the bot
func help(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				&discordgo.MessageEmbed{
					Title: "Game Streams",
					Description: "Game Streams is a bot that keeps track of game announcement streams and can announce when streams are beginning. " +
						"\nUse the `/settings` command to configure the bot to your liking.",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Commands",
							Value:  "`/streams` - List all upcoming streams\n`/help` [admin] - Get help with the bot\n`/settings` [admin] - Change bot settings",
							Inline: false,
						},
						{
							Name: "Settings",
							Value: "Options:\n`channel` the channel for announcing when a stream starts\n`role` the role to ping when a stream starts\nplatforms: enable or disable announcements by platform" +
								"\n\nAll fields are optional, the default settings are to not announce any streams until a channel and one or more platforms are set." +
								"\n\nUse the `/settings` command with no options to see the current settings.",
							Inline: false,
						},
					},
				},
			},
		},
	})
	if respondErr != nil {
		utils.EWLogger.WithPrefix(" CMND").Error("error responding to interaction", "cmd", i.ApplicationCommandData().Name, "err", respondErr)
	}
}

// create an options struct from parsing the options from the interaction and pass it to the settings function
// then respond to the interaction with the updated settings, or an error message if an error occurred
func settings(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := parseOptions(i.ApplicationCommandData().Options)
	var status string
	if *options == (db.Options{}) || options.Reset {
		status = "Current settings:"
	} else {
		status = "Settings successfully updated\nCurrent settings:"
	}
	if options.Reset {
		db.ResetOptions(i.GuildID)
		options = &db.Options{}
	}
	options = db.MergeOptions(i.GuildID, options)
	options.ServerID = i.GuildID

	var channelName string
	var roleName string
	if options.AnnounceChannel != "" {
		channel, cErr := s.Channel(options.AnnounceChannel)
		if cErr != nil {
			utils.EWLogger.WithPrefix(" CMND").Error("error getting channel name", "channel", options.AnnounceChannel, "err", cErr)
			channelName = options.AnnounceChannel
		} else {
			channelName = channel.Name
		}
	}
	if options.AnnounceRole != "" {
		role, rErr := s.State.Role(i.GuildID, options.AnnounceRole)
		if rErr != nil {
			utils.EWLogger.WithPrefix(" CMND").Error("error getting role name", "role", options.AnnounceRole, "err", rErr)
			roleName = options.AnnounceRole
		} else {
			roleName = role.Name
		}
	}
	content := []*discordgo.MessageEmbed{
		&discordgo.MessageEmbed{
			Title:       "Settings",
			Description: status,
			Fields: []*discordgo.MessageEmbedField{
				{},
				{
					Name:   "Announce Channel",
					Value:  utils.PlaceholderText(channelName),
					Inline: false,
				},
				{
					Name:   "Announce Role",
					Value:  utils.PlaceholderText(roleName),
					Inline: false,
				},
				{
					Name:   "Playstation",
					Value:  strconv.FormatBool(options.Playstation),
					Inline: false,
				},
				{
					Name:   "Xbox",
					Value:  strconv.FormatBool(options.Xbox),
					Inline: false,
				},
				{
					Name:   "Nintendo",
					Value:  strconv.FormatBool(options.Nintendo),
					Inline: false,
				},
				{
					Name:   "PC",
					Value:  strconv.FormatBool(options.PC),
					Inline: false,
				},
				{
					Name:   "Awards",
					Value:  strconv.FormatBool(options.Awards),
					Inline: false,
				},
			},
		},
	}
	settingsErr := db.SetOptions(options)
	if settingsErr != nil {
		utils.EWLogger.WithPrefix(" CMND").Error("error setting options", "server", i.GuildID, "err", settingsErr)
		content = []*discordgo.MessageEmbed{
			&discordgo.MessageEmbed{
				Title:       "Settings",
				Description: "An error occurred. Settings have not been updated.",
			},
		}
	}
	respondErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: content,
		},
	})
	if respondErr != nil {
		utils.EWLogger.WithPrefix(" CMND").Error("error responding to interaction", "cmd", i.ApplicationCommandData().Name, "err", respondErr)
	}
}

// parse the options from the interaction into an options struct
func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) *db.Options {
	var o db.Options
	for _, option := range options {
		switch option.Name {
		case "channel":
			o.AnnounceChannel = option.Value.(string)
		case "role":
			o.AnnounceRole = option.Value.(string)
		case "playstation":
			o.Playstation = option.Value.(bool)
		case "xbox":
			o.Xbox = option.Value.(bool)
		case "nintendo":
			o.Nintendo = option.Value.(bool)
		case "pc":
			o.PC = option.Value.(bool)
		case "awards":
			o.Awards = option.Value.(bool)
		case "reset":
			o.Reset = option.Value.(bool)
		}
	}
	return &o
}
