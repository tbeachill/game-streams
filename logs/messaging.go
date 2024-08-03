package logs

import (
	"gamestreams/config"
)

// DM sends a direct message to a user.
func DM(userID string, message string) {
	st, err := Session.UserChannelCreate(userID)
	if err != nil {
		LogError(" MAIN", "error creating DM channel", "err", err)
		return
	}
	_, err = Session.ChannelMessageSend(st.ID, message)
	if err != nil {
		LogError(" MAIN", "error sending DM", "err", err)
	}
}

// DM sends a direct message to the bot owner. The owner's Discord ID is stored in
// the OWNER_ID environment variable.
func DMOwner(message string) {
	DM(config.Values.Discord.OwnerID, message)
}
