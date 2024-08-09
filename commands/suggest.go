package commands

import (
	"github.com/bwmarrin/discordgo"

	"gamestreams/db"
	"gamestreams/logs"
	"gamestreams/utils"
)

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

	embed := &discordgo.MessageEmbed{
		Title:       "Thank you",
		Description: "Your suggestion has been received",
		Color:       0xc3d23e,
	}
	streamName := i.ApplicationCommandData().Options[0].StringValue()
	streamDate := i.ApplicationCommandData().Options[1].StringValue()
	streamURL := i.ApplicationCommandData().Options[2].StringValue()

	suggestion, suggestErr := db.NewSuggestion(streamName, streamDate, streamURL)
	if suggestErr != nil {
		embed = &discordgo.MessageEmbed{
			Title:       "Error",
			Description: suggestErr.Error(),
			Color:       0xc3d23e,
		}
		respond(s, i, embed)
		return
	}

	insertErr := suggestion.Insert()
	if insertErr != nil {
		logs.LogError(" CMND", "error inserting suggestion",
			"err", suggestErr)
		embed = &discordgo.MessageEmbed{
			Title:       "Error",
			Description: "**An error occurred.** Your suggestion may not have been recieved.",
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
