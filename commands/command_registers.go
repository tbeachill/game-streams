package commands

import (
	"github.com/bwmarrin/discordgo"

	"gamestreams/logs"
)

// RegisterCommands registers all commands in the commands slice, which is defined in
// commands.go
func RegisterCommands(appID string, s *discordgo.Session) {
	for _, c := range commands {
		_, err := s.ApplicationCommandCreate(appID, "", c)
		if err != nil {
			logs.LogError(" CMND", "error creating command",
				"cmd", c.Name,
				"err", err)
		}
		logs.LogInfo(" CMND", "registered command", false,
			"cmd", c.Name)
	}
}

// RemoveAllCommands retieves all registered commands and removes them from the
// application.
func RemoveAllCommands(appID string, s *discordgo.Session) {
	commands, err := s.ApplicationCommands(appID, "")
	if err != nil {
		logs.LogError(" CMND", "error removing commands",
			"err", err)
	}
	for _, command := range commands {
		if delErr := s.ApplicationCommandDelete(appID, "", command.ID); delErr != nil {
			logs.LogError(" MAIN", "error removing command",
				"cmd", command.Name,
				"err", delErr)
			continue
		}
		logs.LogInfo(" CMND", "removed command", false,
			"cmd", command.Name)
	}
}

// RegisterHandler registers the command handler functions, defined in handlers.go, for
// each command.
func RegisterHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}
