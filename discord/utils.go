/*
utils.go contains utility functions that are used throughout the bot that are related to
Discord interactions.
*/

package discord

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// GetUserID returns the user ID of the user who sent the interaction.
func GetUserID(i *discordgo.InteractionCreate) string {
	if i.GuildID == "" {
		return i.User.ID
	} else {
		return i.Member.User.ID
	}
}

// GetRoleName takes a role ID and returns a string of the role name.
func GetRoleName(s *discordgo.Session, guildID string, roleID string) string {
	role, err := s.State.Role(guildID, roleID)
	if err != nil {
		return "Role not found"
	}
	return role.Name
}

// DisplayRole ensures that the role name is in the correct format for use in the bot.
func DisplayRole(s *discordgo.Session, guildID string, role string) string {
	// Stop the role from displaying as @@everyone
	if GetRoleName(s, guildID, role) == "@everyone" {
		return "@everyone"
	} else if GetRoleName(s, guildID, role) == "" {
		return ""
	}
	return fmt.Sprintf("<@&%s>", role)
}

// CreateTimestamp creates absolute Discord timestamps from date and time strings.
func CreateTimestamp(d string, t string) (string, string, error) {
	layout := "2006-01-02 15:04"
	if t == "" {
		dt, err := time.Parse(layout, fmt.Sprintf("%s %s", d, "09:00"))
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("<t:%d:d>", dt.Unix()), "TBC", nil
	}
	dt, err := time.Parse(layout, fmt.Sprintf("%s %s", d, t))
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("<t:%d:d>", dt.Unix()), fmt.Sprintf("<t:%d:t>", dt.Unix()), err
}

// CreateTimestampRelative returns a relative Discord timestamp from date and time
// strings. e.g. "in 2 hours"
func CreateTimestampRelative(d string, t string) (string, error) {
	layout := "2006-01-02 15:04"
	dt, err := time.Parse(layout, fmt.Sprintf("%s %s", d, t))
	return fmt.Sprintf("<t:%d:R>", dt.Unix()), err
}
