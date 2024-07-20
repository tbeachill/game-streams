package reports

import (
	"os"

	"github.com/bwmarrin/discordgo"
)

// DM sends a direct message to the bot owner. The owner's Discord ID is stored in
// the OWNER_ID environment variable.
func DM(session *discordgo.Session, message string) {
	st, err := session.UserChannelCreate(os.Getenv("OWNER_ID"))
	if err != nil {
		return
	}
	_, sendErr := session.ChannelMessageSend(st.ID, message)
	if sendErr != nil {
		return
	}
}
