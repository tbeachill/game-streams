/*
suggest.go provides the functions for handling the /suggest command.
*/
package commands

import (
	"github.com/bwmarrin/discordgo"

	"gamestreams/config"
	"gamestreams/db"
	"gamestreams/discord"
	"gamestreams/logs"
)

// suggest allows users to suggest a stream to be added to the database. It extracts the
// stream name, date, and URL from the options then creates a new suggestion. If the
// suggestion is successfully created, it inserts the suggestion into the database and
// responds to the interaction with a success message. If an error occurs, it responds
// with an error message. A suggestion limit is defined in config.toml.
func suggest(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if userIsBlacklisted(i) {
		return
	}
	a := db.CommandData{}
	a.Start(i)
	defer a.End()
	userID := discord.GetUserID(i)
	logs.LogInfo(" CMND", "suggest command", false,
		"user", userID,
		"server", i.GuildID)

	// Check if the user has reached the daily limit for suggestions
	suggestionsToday, countErr := db.CountSuggestions(userID, 1)
	if countErr != nil {
		logs.LogError(" CMND", "error counting suggestions",
			"user", userID,
			"err", countErr)
	}
	if suggestionsToday >= config.Values.Suggestions.DailyLimit {
		embed := &discordgo.MessageEmbed{
			Title:       "Error",
			Description: "You have reached the daily suggestion limit. Try again tomorrow.",
			Color:       config.Values.Discord.EmbedColour,
		}
		respond(s, i, embed)
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Thank you",
		Description: "Your suggestion has been received",
		Color:       config.Values.Discord.EmbedColour,
	}
	streamName := i.ApplicationCommandData().Options[0].StringValue()
	streamDate := i.ApplicationCommandData().Options[1].StringValue()
	streamURL := i.ApplicationCommandData().Options[2].StringValue()

	suggestion, suggestErr := db.NewSuggestion(streamName, streamDate, streamURL)
	if suggestErr != nil {
		embed = &discordgo.MessageEmbed{
			Title:       "Error",
			Description: suggestErr.Error(),
			Color:       config.Values.Discord.EmbedColour,
		}
		respond(s, i, embed)
		return
	}
	suggestion.CommandID = a.CommandID
	insertErr := suggestion.Insert()
	if insertErr != nil {
		logs.LogError(" CMND", "error inserting suggestion",
			"err", insertErr)
		embed = &discordgo.MessageEmbed{
			Title:       "Error",
			Description: "**An error occurred.** Your suggestion may not have been recieved.",
			Color:       config.Values.Discord.EmbedColour,
		}
	}
	respond(s, i, embed)
}

// respond sends a response to the interaction with the provided embed. If an error
// occurs, it logs the error.
func respond(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
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
