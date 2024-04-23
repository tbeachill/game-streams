package reports

import (
	"os"

	"github.com/bwmarrin/discordgo"
)

// Send a direct message to me containing the given message
func DM(session *discordgo.Session, message string) {
	st, err := session.UserChannelCreate(os.Getenv("MY_USER_ID"))
	if err != nil {
		return
	}
	session.ChannelMessageSend(st.ID, message)
}
