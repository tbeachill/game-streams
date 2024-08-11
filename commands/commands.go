package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"gamestreams/config"
	"gamestreams/db"
	"gamestreams/discord"
	"gamestreams/logs"
	"gamestreams/utils"
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

func userIsBlacklisted(i *discordgo.InteractionCreate) bool {
	userID := utils.GetUserID(i)
	blacklisted, reason, expiryDate := db.IsBlacklisted(userID, "user")
	if blacklisted {
		logs.LogInfo(" CMND", "blacklisted user tried to use command", false,
			"user", userID,
			"reason", reason,
			"command", i.ApplicationCommandData().Name)
		discord.DM(userID, fmt.Sprintf("You are blacklisted from using this bot.\n\nReason: %s"+
			"\nExpires: %s ", reason, expiryDate))
		return true
	}
	return false
}

// BlacklistIfSpamming checks if a user is spamming commands and blacklists them if they are.
func BlacklistIfSpamming(i *discordgo.InteractionCreate) {
	userID := utils.GetUserID(i)

	dCount, err := db.CheckUsageByUser(userID, "-1 day")
	if err != nil {
		logs.LogError(" CMND", "error checking command usage",
			"user", userID,
			"err", err)
		return
	}
	hCount, err := db.CheckUsageByUser(userID, "-1 hour")
	if err != nil {
		logs.LogError(" CMND", "error checking command usage",
			"user", userID,
			"err", err)
		return
	}
	if dCount >= config.Values.Blacklist.DailyCommandLimit ||
		hCount >= config.Values.Blacklist.HourlyCommandLimit {
		db.AddToBlacklist(userID, "user", "spamming commands", 2)
	}
}
