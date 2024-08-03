package discord

import (
	"gamestreams/config"
	"gamestreams/logs"
)

// IntroDM sends an introductory DM to a user when they add the bot to their server.
func IntroDM(userID string) {
	message := "🕹 Hello! Thank you for adding me to your server! 🕹\n\n" +
		"To set up your server's announcement channel, announcement role, and which platforms you want to follow, type `/settings` in the server you added me to.\n\n" +
		"For more information, type `/help`."
	logs.LogInfo(" MAIN", "sending intro DM", false, "user", userID)

	DM(userID, message)
}

// DM sends a direct message to a user.
func DM(userID string, message string) {
	st, err := Session.UserChannelCreate(userID)
	if err != nil {
		logs.LogError(" MAIN", "error creating DM channel", "err", err)
		return
	}
	_, err = Session.ChannelMessageSend(st.ID, message)
	if err != nil {
		logs.LogError(" MAIN", "error sending DM", "err", err)
	}
}

// DM sends a direct message to the bot owner. The owner's Discord ID is stored in
// the OWNER_ID environment variable.
func DMOwner(message string) {
	DM(config.Values.Discord.OwnerID, message)
}
