/*
stream_info.go provides the functions for the /streaminfo command.
*/
package commands

import (
	"github.com/bwmarrin/discordgo"

	"gamestreams/config"
	"gamestreams/db"
	"gamestreams/logs"
	"gamestreams/streams"
	"gamestreams/utils"
)

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
				Color:       config.Values.Discord.EmbedColor,
			}
		} else {
			logs.LogError(" CMND", "error creating embeds",
				"err", infoErr)
			embed = &discordgo.MessageEmbed{
				Title:       "Stream Info",
				Description: "An error occurred",
				Color:       config.Values.Discord.EmbedColor,
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
