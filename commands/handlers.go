package commands

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"gamestreams/db"
	"gamestreams/discord"
	"gamestreams/logs"
	"gamestreams/streams"
	"gamestreams/utils"
)

// commandHandlers is a map of command names to their respective handler functions.
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"streams":    listStreams,
	"streaminfo": streamInfo,
	"suggest":    suggest,
	"help":       help,
	"settings":   settings,
}

// listStreams gets a list of all upcoming streams as an embed. If the embed is
// successfully created, it responds to the interaction with the embed. If an error
// occurs, indicating no upcoming streams or an error creating the embed, it responds
// with an error message.
func listStreams(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if userIsBlacklisted(i) {
		return
	}
	a := db.CommandData{}
	a.Start(i)
	defer a.End()

	userID := utils.GetUserID(i)
	logs.LogInfo(" CMND", "list streams command", false,
		"user", userID,
		"server", i.GuildID)

	embed, listErr := streams.StreamList()
	if listErr != nil {
		if listErr.Error() == "no streams found" {
			embed = &discordgo.MessageEmbed{
				Title:       "Upcoming Streams",
				Description: "No streams found",
				Color:       0xc3d23e,
			}
		} else {
			logs.LogError(" CMND", "error creating embeds",
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
		logs.LogError(" CMND", "error responding to interaction",
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
	if userIsBlacklisted(i) {
		return
	}
	a := db.CommandData{}
	a.Start(i)
	defer a.End()

	streamName := i.ApplicationCommandData().Options[0].StringValue()
	userID := utils.GetUserID(i)
	logs.LogInfo(" CMND", "stream info command", false,
		"stream", streamName,
		"user", userID,
		"server", i.GuildID)

	embed, infoErr := streams.StreamInfo(streamName)
	if infoErr != nil {
		if infoErr.Error() == "no streams found" {
			embed = &discordgo.MessageEmbed{
				Title:       "Stream Info",
				Description: "No streams found with that name",
				Color:       0xc3d23e,
			}
		} else {
			logs.LogError(" CMND", "error creating embeds",
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
		logs.LogError(" CMND", "error responding to interaction",
			"cmd", i.ApplicationCommandData().Name,
			"err", respondErr)
	}
}

// suggest allows users to suggest a stream to be added to the database. It extracts the
// stream name, platform, date, time, description, and URL from the options. It then
// creates a stream struct with the information and adds it to the database. If the
// stream is successfully added, it responds to the interaction with a success message.
// If an error occurs, it responds with an error message.
func suggest(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if userIsBlacklisted(i) {
		return
	}
	a := db.CommandData{}
	a.Start(i)
	defer a.End()
	userID := utils.GetUserID(i)
	logs.LogInfo(" CMND", "suggest command", false,
		"user", userID,
		"server", i.GuildID)

	streamName := i.ApplicationCommandData().Options[0].StringValue()
	streamDate := i.ApplicationCommandData().Options[1].StringValue()
	streamURL := i.ApplicationCommandData().Options[2].StringValue()

	suggestion := db.Suggestion{
		Name: streamName,
		Date: streamDate,
		URL:  streamURL,
	}

	suggestErr := suggestion.Insert()
	if suggestErr != nil {
		logs.LogError(" CMND", "error inserting suggestion",
			"err", suggestErr)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Thank you",
		Description: "Your suggestion has been received",
		Color:       0xc3d23e,
	}
	respondErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if respondErr != nil {
		logs.LogError(" CMND", "error responding to interaction",
			"cmd", i.ApplicationCommandData().Name,
			"err", respondErr)
	}
}

// help responds to the interaction with a help message containing information about
// the bot and its commands. This command is only available to server administrators.
func help(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if userIsBlacklisted(i) {
		return
	}
	a := db.CommandData{}
	a.Start(i)
	defer a.End()

	userID := utils.GetUserID(i)
	logs.LogInfo(" CMND", "help command", false,
		"user", userID,
		"server", i.GuildID)

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
		logs.LogError(" CMND", "error responding to interaction",
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
	if userIsBlacklisted(i) {
		return
	}
	a := db.CommandData{}
	a.Start(i)
	defer a.End()

	userID := utils.GetUserID(i)
	logs.LogInfo(" CMND", "settings command", false,
		"user", userID,
		"server", i.GuildID)

	options := parseOptions(i.ApplicationCommandData().Options)

	var status string
	if *options == (db.Settings{}) || options.Reset {
		status = "Current settings:"
	} else {
		status = "Settings successfully updated\nCurrent settings:"
	}
	if options.Reset {
		*options = db.NewSettings(i.GuildID)
		if optErr := options.Set(); optErr != nil {
			logs.LogError(" CMND", "error resetting options",
				"server", i.GuildID,
				"err", optErr)

			status = "An error occurred. Settings may not have been reset."
		}
	}
	var currentOptions = db.NewSettings(i.GuildID)

	if getOptErr := currentOptions.Get(i.GuildID); getOptErr != nil {
		logs.LogError(" CMND", "error getting options",
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

			logs.LogError(" CMND", "error getting channel name",
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
			logs.LogError(" CMND", "error getting role name",
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
		logs.LogError(" CMND", "error setting options",
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
		logs.LogError(" CMND", "error responding to interaction",
			"cmd", i.ApplicationCommandData().Name,
			"err", respondErr)
	}
}

// parseOptions parses the options from the interaction into a settings struct.
func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) *db.Settings {
	var s db.Settings
	for _, option := range options {
		switch option.Name {
		case "channel":
			s.AnnounceChannel.Value = option.Value.(string)
			s.AnnounceChannel.Set = true
		case "role":
			s.AnnounceRole.Value = option.Value.(string)
			s.AnnounceRole.Set = true
		case "playstation":
			s.Playstation.Value = option.BoolValue()
			s.Playstation.Set = true
		case "xbox":
			s.Xbox.Value = option.BoolValue()
			s.Xbox.Set = true
		case "nintendo":
			s.Nintendo.Value = option.BoolValue()
			s.Nintendo.Set = true
		case "pc":
			s.PC.Value = option.BoolValue()
			s.PC.Set = true
		case "vr":
			s.VR.Value = option.BoolValue()
			s.VR.Set = true
		case "reset":
			s.Reset = option.BoolValue()
		}
	}
	return &s
}

func userIsBlacklisted(i *discordgo.InteractionCreate) bool {
	userID := utils.GetUserID(i)
	blacklisted, reason := db.IsBlacklisted(userID, "user")
	if blacklisted {
		logs.LogInfo(" CMND", "blacklisted user tried to use command", false,
			"user", userID,
			"reason", reason,
			"command", i.ApplicationCommandData().Name)
		discord.DM(userID, fmt.Sprintf("You are blacklisted from using this bot. Reason: %s", reason))
		return true
	}
	return false
}
