package commands

import (
	"github.com/bwmarrin/discordgo"

	"gamestreambot/streams"
	"gamestreambot/utils"
)

// map of command names to their respective functions
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"list-streams": listStreams,
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
