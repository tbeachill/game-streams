package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"gamestreams/config"
	"gamestreams/db"
	"gamestreams/logs"
	"gamestreams/utils"
)

// help responds with a help message for the bot.
// Help messages are specific to the command requested. If no command is requested,
// a general help message is sent.
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

	var content []*discordgo.MessageEmbed
	if len(i.ApplicationCommandData().Options) == 0 {
		content = helpGeneral()
	} else {
		switch i.ApplicationCommandData().Options[0].Value {
		case "streams":
			content = helpStreams()
		case "streaminfo":
			content = helpStreamInfo()
		case "suggest":
			content = helpSuggest()
		case "settings":
			content = helpSettings()
		default:
			content = helpGeneral()
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

// helpGeneral returns a general help message for the bot.
func helpGeneral() []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{
		{
			Title: "Game Streams",
			Description: "Game Streams is a bot that keeps track of game announcement streams " +
				"and can announce when streams are beginning. \n\nUse the `/settings` command in your server " +
				"to configure the bot to your liking.",
			Color: 0xc3d23e,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "Commands",
					Value: "`/streams` - List all upcoming streams" +
						"\n`/streaminfo` - Get information on a specific stream by title" +
						"\n`/suggest` - Suggest a stream to be added to the database" +
						"\n`/help` - Get help with the bot" +
						"\n`/settings` [admin] - Change bot settings",
					Inline: false,
				},
				{
					Name: "Documents",
					Value: fmt.Sprintf("Privacy Policy: %s\n", config.Values.Documents.PrivacyPolicy) +
						fmt.Sprintf("Terms of Service: %s\n", config.Values.Documents.TermsOfService) +
						fmt.Sprintf("Changelog: %s\n", config.Values.Documents.Changelog),
					Inline: false,
				},
				{
					Name: "Information",
					Value: fmt.Sprintf("Version: %s\n", config.Values.Bot.Version) +
						fmt.Sprintf("Release Date: %s\n", config.Values.Bot.ReleaseDate),
				},
			},
		},
	}
}

// helpStreams returns a help message for the /streams command.
func helpStreams() []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{
		{
			Title: "/streams",
			Description: "List all upcoming streams. Streams are sorted by date and time." +
				"\n\nStreams that have already started will not be listed. " +
				fmt.Sprintf("\n\nList is limited to %d streams.", config.Values.Streams.Limit),
			Color: 0xc3d23e,
		},
	}
}

// helpStreamInfo returns a help message for the /streaminfo command.
func helpStreamInfo() []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{
		{
			Title: "/streaminfo",
			Description: "Get information on a specific stream by title." +
				"\n\nThe name field is required.",
			Color: 0xc3d23e,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "name",
					Value:  "The name of the stream to search for, partial matches are allowed.",
					Inline: false,
				},
			},
		},
	}
}

// helpSuggest returns a help message for the /suggest command.
func helpSuggest() []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{
		{
			Title: "/suggest",
			Description: "Suggest a stream to be added to the database. Suggestions will be reviewed by " +
				"the bot owner and added to the database if they are valid.\n\nAll fields are required.",
			Color: 0xc3d23e,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "name",
					Value:  "The name of the stream to suggest",
					Inline: false,
				},
				{
					Name:   "date",
					Value:  "The date of the stream in the format `YYYY-MM-DD`",
					Inline: false,
				},
				{
					Name:   "url",
					Value:  "The URL of the stream",
					Inline: false,
				},
			},
		},
	}
}

// helpSettings returns a help message for the /settings command.
func helpSettings() []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{
		{
			Title: "/settings",
			Description: "Settings for the Game Streams bot. These need to be set in your server to " +
				"enable the bot to announce streams.\n\nOnly server administrators can use this command." +
				"\n\nAll fields are optional.",
			Color: 0xc3d23e,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "channel",
					Value: "The channel for announcing when a stream starts. " +
						"**If not set, the bot will not announce streams.**",
					Inline: false,
				},
				{
					Name: "role",
					Value: "The role to ping when a stream starts. " +
						"**If not set, the bot will still announce streams but will not ping anyone.**",
					Inline: false,
				},
				{
					Name:   "reset",
					Value:  "Reset all settings to default. Use `True` to reset.",
					Inline: false,
				},
				{
					Name: "platforms",
					Value: "Enable or disable announcements by platform. " +
						"Use `True` to enable and `False` to disable annoucements.",
					Inline: false,
				},
			},
		},
	}
}
