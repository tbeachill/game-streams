/*
streams.go contains functions for handling the /streams command.
*/
package commands

import (
	"github.com/bwmarrin/discordgo"

	"gamestreams/db"
	"gamestreams/logs"
	"gamestreams/streams"
	"gamestreams/utils"
)

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
