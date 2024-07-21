package commands

import (
	"strconv"

	"github.com/bwmarrin/discordgo"

	"gamestreambot/db"
	"gamestreambot/streams"
	"gamestreambot/utils"
)

// commandHandlers is a map of command names to their respective handler functions.
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"streams":    listStreams,
	"streaminfo": streamInfo,
	"help":       help,
	"settings":   settings,
}

// listStreams gets a list of all upcoming streams as an embed. If the embed is
// successfully created, it responds to the interaction with the embed. If an error
// occurs, indicating no upcoming streams or an error creating the embed, it responds
// with an error message.
func listStreams(s *discordgo.Session, i *discordgo.InteractionCreate) {
	blacklisted, _ := db.IsBlacklisted(i.User.ID, "user")
	if blacklisted {
		utils.LogInfo(" CMND", "command ran by blacklisted user", true,
			"cmd", i.ApplicationCommandData().Name,
			"user_id", i.User.ID,
			"name", i.User.Username)
		return
	}
	embed, listErr := streams.StreamList()
	if listErr != nil {
		if listErr.Error() == "no streams found" {
			embed = &discordgo.MessageEmbed{
				Title:       "Upcoming Streams",
				Description: "No streams found",
				Color:       0xc3d23e,
			}
		} else {
			utils.LogError(" CMND", "error creating embeds",
				"err", listErr)
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
		utils.LogError(" CMND", "error responding to interaction",
			"cmd", i.ApplicationCommandData().Name,
			"err", respondErr)
	}
}

// streamInfo gets the information for a specific stream by title. It extracts the
// stream name from the options then gets the stream information from the database.
// If the stream is found, it creates an embed with the stream information and responds
// to the interaction with the embed. If the stream is not found or an error occurs,
// it responds with an error message.
func streamInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	blacklisted, _ := db.IsBlacklisted(i.User.ID, "user")
	if blacklisted {
		utils.LogInfo(" CMND", "command ran by blacklisted user", true,
			"cmd", i.ApplicationCommandData().Name,
			"user_id", i.User.ID,
			"name", i.User.Username)
		return
	}
	streamName := i.ApplicationCommandData().Options[0].Value.(string)
	embed, infoErr := streams.StreamInfo(streamName)
	if infoErr != nil {
		if infoErr.Error() == "no streams found" {
			embed = &discordgo.MessageEmbed{
				Title:       "Stream Info",
				Description: "No streams found with that name",
				Color:       0xc3d23e,
			}
		} else {
			utils.LogError(" CMND", "error creating embeds",
				"err", infoErr)
			embed = &discordgo.MessageEmbed{
				Title:       "Stream Info",
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
		utils.LogError(" CMND", "error responding to interaction",
			"cmd", i.ApplicationCommandData().Name,
			"err", respondErr)
	}
}

// help responds to the interaction with a help message containing information about
// the bot and its commands. This command is only available to server administrators.
func help(s *discordgo.Session, i *discordgo.InteractionCreate) {
	blacklisted, _ := db.IsBlacklisted(i.User.ID, "user")
	if blacklisted {
		utils.LogInfo(" CMND", "command ran by blacklisted user", true,
			"cmd", i.ApplicationCommandData().Name,
			"user_id", i.User.ID,
			"name", i.User.Username)
		return
	}
	respondErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				&discordgo.MessageEmbed{
					Title: "Game Streams",
					Description: "Game Streams is a bot that keeps track of game announcement streams and can announce when streams are beginning. " +
						"\n\nUse the `/settings` command in your server to configure the bot to your liking.",
					Color: 0xc3d23e,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Commands",
							Value:  "`/streams` - List all upcoming streams\n`/streaminfo` - Get information on a specific stream by title\n`/help` [admin] - Get help with the bot\n`/settings` [admin] - Change bot settings **(server only - can't use in DMs)**",
							Inline: false,
						},
						{
							Name: "Settings",
							Value: "Options:\n`channel` the channel for announcing when a stream starts\n`role` the role to ping when a stream starts\nplatforms: enable or disable announcements by platform" +
								"\n\n- All fields are optional, the default settings are to not announce any streams until a channel and one or more platforms are set.\n" +
								"- If no channel or no platforms are selected, I will not announce streams.\n" +
								"- If no role is selected, I will still announce streams but will not ping anyone.\n" +
								"- The platform settings only control which streams are announced, not which streams are listed by the `/streams` command.\n\n" +
								"Use the `/settings` command with no options to view the current settings.\n\n" +
								"Use the `/settings` command with the `reset` option as `True` to reset all settings to default.",
							Inline: false,
						},
					},
				},
			},
		},
	})
	if respondErr != nil {
		utils.LogError(" CMND", "error responding to interaction",
			"cmd", i.ApplicationCommandData().Name,
			"err", respondErr)
	}
}

// settings updates the bot settings for the server. It parses the options from the
// interaction into an options struct. If the options are empty, it responds with the
// current settings. If the reset option is true, it resets the settings to default.
// If the options are not empty, it first gets the current settings from the database,
// then merges the new settings with the current settings. It then writes the new
// settings to the database and responds with the updated settings. If an error occurs,
// it responds with an error message.
func settings(s *discordgo.Session, i *discordgo.InteractionCreate) {
	blacklisted, _ := db.IsBlacklisted(i.User.ID, "user")
	if blacklisted {
		utils.LogInfo(" CMND", "command ran by blacklisted user", true,
			"cmd", i.ApplicationCommandData().Name,
			"user_id", i.User.ID,
			"name", i.User.Username)
		return
	}
	options := parseOptions(i.ApplicationCommandData().Options)
	utils.LogInfo(" CMND", "options", false, "options", options)

	var status string
	if *options == (db.Options{}) || options.Reset {
		status = "Current settings:"
	} else {
		status = "Settings successfully updated\nCurrent settings:"
	}
	if options.Reset {
		*options = db.NewOptions(i.GuildID)
		if optErr := options.Set(); optErr != nil {
			utils.LogError(" CMND", "error resetting options",
				"server", i.GuildID,
				"err", optErr)

			status = "An error occurred. Settings may not have been reset."
		}
	}
	var currentOptions = db.NewOptions(i.GuildID)

	if getOptErr := currentOptions.Get(i.GuildID); getOptErr != nil {
		utils.LogError(" CMND", "error getting options",
			"server", i.GuildID,
			"err", getOptErr)

		status = "An error occurred. Settings may have not been updated."
	}
	currentOptions.Merge(*options)

	var channelName string
	var roleName string
	if currentOptions.AnnounceChannel.Value != "" {
		channel, cErr := s.Channel(currentOptions.AnnounceChannel.Value)
		if cErr != nil {

			utils.LogError(" CMND", "error getting channel name",
				"channel", currentOptions.AnnounceChannel,
				"err", cErr)

			channelName = currentOptions.AnnounceChannel.Value
		} else {
			channelName = channel.Name
		}
	}
	if currentOptions.AnnounceRole.Value != "" {
		role, rErr := s.State.Role(i.GuildID, currentOptions.AnnounceRole.Value)
		if rErr != nil {
			utils.LogError(" CMND", "error getting role name",
				"role", currentOptions.AnnounceRole,
				"err", rErr)

			roleName = currentOptions.AnnounceRole.Value
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
					Value:  strconv.FormatBool(currentOptions.Playstation.Value),
					Inline: false,
				},
				{
					Name:   "Xbox",
					Value:  strconv.FormatBool(currentOptions.Xbox.Value),
					Inline: false,
				},
				{
					Name:   "Nintendo",
					Value:  strconv.FormatBool(currentOptions.Nintendo.Value),
					Inline: false,
				},
				{
					Name:   "PC",
					Value:  strconv.FormatBool(currentOptions.PC.Value),
					Inline: false,
				},
				{
					Name:   "VR",
					Value:  strconv.FormatBool(currentOptions.VR.Value),
					Inline: false,
				},
			},
		},
	}
	settingsErr := currentOptions.Set()
	if settingsErr != nil {
		utils.LogError(" CMND", "error setting options",
			"server", i.GuildID,
			"err", settingsErr)

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
		utils.LogError(" CMND", "error responding to interaction",
			"cmd", i.ApplicationCommandData().Name,
			"err", respondErr)
	}
}

// parseOptions parses the options from the interaction into an options struct.
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
		case "nintendo":
			o.Nintendo.Value = option.Value.(bool)
			o.Nintendo.Set = true
		case "pc":
			o.PC.Value = option.Value.(bool)
			o.PC.Set = true
		case "vr":
			o.VR.Value = option.Value.(bool)
			o.VR.Set = true
		case "reset":
			o.Reset = option.Value.(bool)
		}
	}
	return &o
}
