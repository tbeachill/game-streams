package commands

import (
	"github.com/bwmarrin/discordgo"

	"gamestreambot/db"
	"gamestreambot/streams"
	"gamestreambot/utils"
)

// map of command names to their respective functions
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"list-streams": listStreams,
	"settings":     settings,
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

// create an options struct from parsing the options from the interaction and pass it to the settings function
func settings(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := parseOptions(i.ApplicationCommandData().Options)
	options.ServerID = i.GuildID

	settingsErr := db.SetOptions(options)
	if settingsErr != nil {
		utils.EWLogger.WithPrefix(" CMND").Error("error setting options", "server", i.GuildID, "err", settingsErr)
	}

	respondErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Settings updated.",
		},
	})
	if respondErr != nil {
		utils.EWLogger.WithPrefix(" CMND").Error("error responding to interaction", "cmd", i.ApplicationCommandData().Name, "err", respondErr)
	}
}

// parse the options from the interaction into an options struct
func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) *db.Options {
	var o db.Options
	for _, option := range options {
		switch option.Name {
		case "channel":
			o.AnnounceChannel = option.Value.(string)
		case "role":
			o.AnnounceRole = option.Value.(string)
		case "playstation":
			o.Playstation = option.Value.(bool)
		case "xbox":
			o.Xbox = option.Value.(bool)
		case "nintendo":
			o.Nintendo = option.Value.(bool)
		case "pc":
			o.PC = option.Value.(bool)
		case "awards":
			o.Awards = option.Value.(bool)
		}
	}
	return &o
}
