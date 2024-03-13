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
	embed, listErr := streams.StreamList()
	if listErr != nil {
		if listErr.Error() == "no streams found" {
			embed = &discordgo.MessageEmbed{
				Title:       "Upcoming Streams",
				Description: "No streams found",
				Color:       0xc3d23e,
			}
		} else {
			utils.Log.ErrorWarn.WithPrefix(" CMND").Error("error creating embeds", "err", listErr)
			embed = &discordgo.MessageEmbed{
				Title:       "Upcoming Streams",
				Description: "An error occurred",
				Color:       0xc3d23e,
			}
		}
	}

	respondErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if respondErr != nil {
		utils.Log.ErrorWarn.WithPrefix(" CMND").Error("error responding to interaction", "cmd", i.ApplicationCommandData().Name, "err", respondErr)
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
						"\n\nUse the `/settings` command to configure the bot to your liking.",
					Color: 0xc3d23e,
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
		utils.Log.ErrorWarn.WithPrefix(" CMND").Error("error responding to interaction", "cmd", i.ApplicationCommandData().Name, "err", respondErr)
	}
}

// create an options struct from parsing the options from the interaction and pass it to the settings function
// then respond to the interaction with the updated settings, or an error message if an error occurred
func settings(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := parseOptions(i.ApplicationCommandData().Options)
	utils.Log.Info.WithPrefix(" CMND").Info("options", "options", options)
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
	if options.AnnounceChannel.Value != "" {
		channel, cErr := s.Channel(options.AnnounceChannel.Value)
		if cErr != nil {
			utils.Log.ErrorWarn.WithPrefix(" CMND").Error("error getting channel name", "channel", options.AnnounceChannel, "err", cErr)
			channelName = options.AnnounceChannel.Value
		} else {
			channelName = channel.Name
		}
	}
	if options.AnnounceRole.Value != "" {
		role, rErr := s.State.Role(i.GuildID, options.AnnounceRole.Value)
		if rErr != nil {
			utils.Log.ErrorWarn.WithPrefix(" CMND").Error("error getting role name", "role", options.AnnounceRole, "err", rErr)
			roleName = options.AnnounceRole.Value
		} else {
			roleName = role.Name
		}
	}
	content := []*discordgo.MessageEmbed{
		&discordgo.MessageEmbed{
			Title:       "Settings",
			Description: status,
			Color:       0xc3d23e,
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
					Value:  strconv.FormatBool(options.Playstation.Value),
					Inline: false,
				},
				{
					Name:   "Xbox",
					Value:  strconv.FormatBool(options.Xbox.Value),
					Inline: false,
				},
				{
					Name:   "Nintendo",
					Value:  strconv.FormatBool(options.Nintendo.Value),
					Inline: false,
				},
				{
					Name:   "PC",
					Value:  strconv.FormatBool(options.PC.Value),
					Inline: false,
				},
				{
					Name:   "VR",
					Value:  strconv.FormatBool(options.VR.Value),
					Inline: false,
				},
				{
					Name:   "Awards",
					Value:  strconv.FormatBool(options.Awards.Value),
					Inline: false,
				},
			},
		},
	}
	settingsErr := db.SetOptions(options)
	if settingsErr != nil {
		utils.Log.ErrorWarn.WithPrefix(" CMND").Error("error setting options", "server", i.GuildID, "err", settingsErr)
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
		utils.Log.ErrorWarn.WithPrefix(" CMND").Error("error responding to interaction", "cmd", i.ApplicationCommandData().Name, "err", respondErr)
	}
}

// parse the options from the interaction into an options struct
func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) *db.Options {
	var o db.Options
	for _, option := range options {
		switch option.Name {
		case "channel":
			o.AnnounceChannel.Value = option.Value.(string)
			o.AnnounceChannel.Set = true
		case "role":
			o.AnnounceRole.Value = option.Value.(string)
			o.AnnounceRole.Set = true
		case "playstation":
			o.Playstation.Value = option.Value.(bool)
			o.Playstation.Set = true
		case "xbox":
			o.Xbox.Value = option.Value.(bool)
			o.Xbox.Set = true
			utils.Log.Info.Info("xbox", "value", option.Value)
		case "nintendo":
			o.Nintendo.Value = option.Value.(bool)
			o.Nintendo.Set = true
		case "pc":
			o.PC.Value = option.Value.(bool)
			o.PC.Set = true
		case "vr":
			o.VR.Value = option.Value.(bool)
			o.VR.Set = true
		case "awards":
			o.Awards.Value = option.Value.(bool)
			o.Awards.Set = true
		case "reset":
			o.Reset = option.Value.(bool)
		}
	}
	return &o
}
