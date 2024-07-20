package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"gamestreambot/reports"
	"gamestreambot/utils"
)

// RegisterCommands registers all commands in the commands slice, which is defined in
// commands.go
func RegisterCommands(appID string, s *discordgo.Session) {
	for _, c := range commands {
		_, err := s.ApplicationCommandCreate(appID, "", c)
		if err != nil {
			utils.Log.ErrorWarn.WithPrefix(" MAIN").Error("error creating command",
				"cmd", c.Name,
				"err", err)

			reports.DM(utils.Session, fmt.Sprintf("error creating command:\n\tcmd=%s\n\terr=%s",
				c.Name,
				err))
		}
		utils.Log.Info.WithPrefix(" MAIN").Info("registered command",
			"cmd", c.Name)
	}
}

// RemoveAllCommands retieves all registered commands and removes them from the
// application.
func RemoveAllCommands(appID string, s *discordgo.Session) {
	commands, err := s.ApplicationCommands(appID, "")
	if err != nil {
		utils.Log.ErrorWarn.WithPrefix(" MAIN").Error("error removing commands",
			"err", err)

		reports.DM(utils.Session, fmt.Sprintf("error removing commands:\n\terr=%s",
			err))
	}
	for _, command := range commands {
		if delErr := s.ApplicationCommandDelete(appID, "", command.ID); delErr != nil {

			utils.Log.ErrorWarn.WithPrefix(" MAIN").Error("error removing command",
				"cmd", command.Name,
				"err", delErr)

			reports.DM(utils.Session, fmt.Sprintf("error removing command:\n\tcmd=%s\n\terr=%s",
				command.Name,
				delErr))
			continue
		}
		utils.Log.Info.WithPrefix(" MAIN").Info("removed command", "cmd", command.Name)
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
