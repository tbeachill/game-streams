/*
blacklist.go contains functions related to blacklisting users and servers from using
the bot.
*/
package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/config"
	"gamestreams/db"
	"gamestreams/discord"
	"gamestreams/logs"
)

// userIsBlacklisted checks if a user is blacklisted from using the bot.
// It returns true if the user is blacklisted and false if they are not.
//
// If the user is blacklisted, it sends a DM to the user with the reason for the
// blacklist and the date the blacklist expires. It then updates the last messaged
// field in the database to the current date so that the user is not spammed with
// messages.
func userIsBlacklisted(i *discordgo.InteractionCreate) bool {
	userID := discord.GetUserID(i)
	blacklisted, b := db.IsBlacklisted(userID)

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
		if b.LastMessaged == "" ||
			time.Now().Compare(lastMessaged) >= config.Values.Blacklist.DaysBetweenMessages {
			discord.DM(userID, fmt.Sprintf("You are blacklisted from using this bot.\n\n"+
				"**Reason:** `%s`\n**Expires:** `%s`",
				b.Reason, b.DateExpires))
			db.UpdateLastMessaged(userID)
		}
		return true
	}
	return false
}

// BlacklistIfSpamming checks if a user is spamming commands and exceeding the daily
// or hourly command limits as specified in config.toml. If the user is spamming
// commands, it adds them to the blacklist with a reason of "spamming commands"
// and a duration of 2 days.
func BlacklistIfSpamming(i *discordgo.InteractionCreate) {
	userID := discord.GetUserID(i)

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
