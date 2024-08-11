package commands

import (
	"github.com/bwmarrin/discordgo"

	"gamestreams/db"
)

// commandHandlers is a map of command names to their respective handler functions.
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"streams":    listStreams,
	"streaminfo": streamInfo,
	"suggest":    suggest,
	"help":       help,
	"settings":   settings,
}

// parseOptions parses the options from the interaction into a settings struct.
func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) *db.Settings {
	var s db.Settings
	for _, option := range options {
		switch option.Name {
		case "channel":
			s.AnnounceChannel.Value = option.Value.(string)
			s.AnnounceChannel.Set = true
		case "role":
			s.AnnounceRole.Value = option.Value.(string)
			s.AnnounceRole.Set = true
		case "playstation":
			s.Playstation.Value = option.BoolValue()
			s.Playstation.Set = true
		case "xbox":
			s.Xbox.Value = option.BoolValue()
			s.Xbox.Set = true
		case "nintendo":
			s.Nintendo.Value = option.BoolValue()
			s.Nintendo.Set = true
		case "pc":
			s.PC.Value = option.BoolValue()
			s.PC.Set = true
		case "vr":
			s.VR.Value = option.BoolValue()
			s.VR.Set = true
		case "reset":
			s.Reset = option.BoolValue()
		}
	}
	return &s
}
