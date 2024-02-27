package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"gamestreambot/streams"
)

// map of command names to their respective functions
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"list-streams": listStreams,
}

// list all upcoming streams
func listStreams(s *discordgo.Session, i *discordgo.InteractionCreate) {
	content, getErr := streams.StreamList()
	if getErr != nil {
		log.Printf("error getting stream list: %e\n", getErr)
		content = "An error occurred."
	}

	respondErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if respondErr != nil {
		log.Printf("error responding to interaction %s: %e\n", i.ApplicationCommandData().Name, respondErr)
	}
}
