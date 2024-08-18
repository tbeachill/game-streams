/*
messages.go provides functions for sending messages to users and the bot owner.
*/
package logs

import (
	"gamestreams/config"
)

// DM sends a direct message to a user.
func DM(userID string, message string) {
	st, err := Session.UserChannelCreate(userID)
	if err != nil {
		LogError(" LOGS", "error creating DM channel", "err", err)
		return
	}
	_, err = Session.ChannelMessageSend(st.ID, message)
	if err != nil {
		LogError(" LOGS", "error sending DM", "err", err)
	}
}

// DM sends a direct message to the bot owner. The owner's Discord ID is set in
// config.toml.
func DMOwner(message string) {
	DM(config.Values.Discord.OwnerID, message)
}
