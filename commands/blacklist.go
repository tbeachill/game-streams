package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/config"
	"gamestreams/db"
	"gamestreams/discord"
	"gamestreams/logs"
	"gamestreams/utils"

)

// userIsBlacklisted checks if a user is blacklisted from using the bot.
// If the user is blacklisted, it sends a DM to the user with the reason for the blacklist.
func userIsBlacklisted(i *discordgo.InteractionCreate) bool {
	userID := utils.GetUserID(i)
	blacklisted, b := db.IsBlacklisted(userID, "user")

	if blacklisted {
		logs.LogInfo(" CMND", "blacklisted user tried to use command", false,
			"user", userID,
			"reason", b.Reason,
			"command", i.ApplicationCommandData().Name)
		lastMessaged, timeErr := time.Parse("2006-01-02", b.LastMessaged)
		if timeErr != nil {
			logs.LogError(" CMND", "error parsing time",
				"user", userID,
				"err", timeErr)
		}
		if b.LastMessaged == "" || time.Now().Compare(lastMessaged) > 0 {
			discord.DM(userID, fmt.Sprintf("You are blacklisted from using this bot.\n\nReason: %s"+
				"\nExpires: %s ", b.Reason, b.DateExpires))
			db.UpdateLastMessaged(userID)
		}
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
