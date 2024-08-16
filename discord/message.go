/*
message.go contains functions that send messages to users.
*/
package discord

import (
	"gamestreams/config"
	"gamestreams/logs"
)

// IntroDM sends an introductory DM to a server owner when the bot is added to a server.
func IntroDM(userID string) {
	message := "🕹 Hello! Thank you for adding me to your server! 🕹\n\n" +
		"To set up the bot to announce when streams are starting, and which platforms you" +
		" want to follow, type `/settings` in the server you added me to.\n\nFor help" +
		" with the bot and its commands, type `/help`. Commands can only be used" +
		" in servers."
	logs.LogInfo("DSCRD", "sending intro DM", false, "user", userID)

	DM(userID, message)
}

// DM sends a direct message containing the given message to the user with the given ID.
func DM(userID string, message string) {
	st, err := Session.UserChannelCreate(userID)
	if err != nil {
		logs.LogError("DSCRD", "error creating DM channel", "err", err)
		return
	}
	_, err = Session.ChannelMessageSend(st.ID, message)
	if err != nil {
		logs.LogError("DSCRD", "error sending DM", "err", err)
	}
}

// DM sends a direct message to the bot owner. The owner ID is set in config.toml.
func DMOwner(message string) {
	DM(config.Values.Discord.OwnerID, message)
}
