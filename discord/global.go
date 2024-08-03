package discord

import "github.com/bwmarrin/discordgo"

// Session is a pointer to the discord session.
var Session *discordgo.Session

// RegisterSession sets the global Session variable.
func RegisterSession(s *discordgo.Session) {
	Session = s
}
