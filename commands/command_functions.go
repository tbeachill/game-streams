package commands

import (
	"github.com/bwmarrin/discordgo"

	"gamestreambot/utils"
)

// register commands listed in commands.go
func RegisterCommands(appID string, s *discordgo.Session) {
	for _, c := range commands {
		_, err := s.ApplicationCommandCreate(appID, "", c)
		if err != nil {
			utils.EWLogger.WithPrefix(" CMND").Error("error creating command", "cmd", c.Name, "err", err)
		}
	}
}

// remove all registered global commands
func RemoveAllCommands(appID string, s *discordgo.Session) {
	commands, err := s.ApplicationCommands(appID, "")
	if err != nil {
		utils.EWLogger.WithPrefix(" CMND").Error("error removing commands", "err", err)
	}
	for _, command := range commands {
		s.ApplicationCommandDelete(appID, "", command.ID)
	}
}

// register command handlers in handlers.go
func RegisterHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}
