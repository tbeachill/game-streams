package commands

import (
	"strconv"

	"github.com/bwmarrin/discordgo"

	"gamestreams/db"
	"gamestreams/logs"
	"gamestreams/utils"
)

// settings updates the bot settings for the server. It parses the options from the
// interaction into an options struct. If the options are empty, it responds with the
// current settings. If the reset option is true, it resets the settings to default.
// If the options are not empty, it first gets the current settings from the database,
// then merges the new settings with the current settings. It then writes the new
// settings to the database and responds with the updated settings. If an error occurs,
// it responds with an error message.
func settings(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if userIsBlacklisted(i) {
		return
	}
	a := db.CommandData{}
	a.Start(i)
	defer a.End()

	userID := utils.GetUserID(i)
	logs.LogInfo(" CMND", "settings command", false,
		"user", userID,
		"server", i.GuildID)

	options := parseOptions(i.ApplicationCommandData().Options)

	var status string
	if *options == (db.Settings{}) || options.Reset {
		status = "Current settings:"
	} else {
		status = "Settings successfully updated\nCurrent settings:"
	}
	if options.Reset {
		*options = db.NewSettings(i.GuildID)
		if optErr := options.Set(); optErr != nil {
			logs.LogError(" CMND", "error resetting options",
				"server", i.GuildID,
				"err", optErr)

			status = "An error occurred. Settings may not have been reset."
		}
	}
	var currentOptions = db.NewSettings(i.GuildID)

	if getOptErr := currentOptions.Get(i.GuildID); getOptErr != nil {
		logs.LogError(" CMND", "error getting options",
			"server", i.GuildID,
			"err", getOptErr)

		status = "An error occurred. Settings may have not been updated."
	}
	currentOptions.Merge(*options)

	var channelName string
	var roleName string
	if currentOptions.AnnounceChannel.Value != "" {
		channel, cErr := s.Channel(currentOptions.AnnounceChannel.Value)
		if cErr != nil {

			logs.LogError(" CMND", "error getting channel name",
				"channel", currentOptions.AnnounceChannel,
				"err", cErr)

			channelName = currentOptions.AnnounceChannel.Value
		} else {
			channelName = channel.Name
		}
	}
	if currentOptions.AnnounceRole.Value != "" {
		role, rErr := s.State.Role(i.GuildID, currentOptions.AnnounceRole.Value)
		if rErr != nil {
			logs.LogError(" CMND", "error getting role name",
				"role", currentOptions.AnnounceRole,
				"err", rErr)

			roleName = currentOptions.AnnounceRole.Value
		} else {
			roleName = role.Name
		}
	}
	content := []*discordgo.MessageEmbed{
		&discordgo.MessageEmbed{
			Title:       "Settings",
			Description: status,
			Color:       0xc3d23e,
			Fields: []*discordgo.MessageEmbedField{
				{},
				{
					Name:   "Announce Channel",
					Value:  utils.PlaceholderText(channelName),
					Inline: false,
				},
				{
					Name:   "Announce Role",
					Value:  utils.PlaceholderText(roleName),
					Inline: false,
				},
				{
					Name:   "Playstation",
					Value:  strconv.FormatBool(currentOptions.Playstation.Value),
					Inline: false,
				},
				{
					Name:   "Xbox",
					Value:  strconv.FormatBool(currentOptions.Xbox.Value),
					Inline: false,
				},
				{
					Name:   "Nintendo",
					Value:  strconv.FormatBool(currentOptions.Nintendo.Value),
					Inline: false,
				},
				{
					Name:   "PC",
					Value:  strconv.FormatBool(currentOptions.PC.Value),
					Inline: false,
				},
				{
					Name:   "VR",
					Value:  strconv.FormatBool(currentOptions.VR.Value),
					Inline: false,
				},
			},
		},
	}
	settingsErr := currentOptions.Set()
	if settingsErr != nil {
		logs.LogError(" CMND", "error setting options",
			"server", i.GuildID,
			"err", settingsErr)

		content = []*discordgo.MessageEmbed{
			&discordgo.MessageEmbed{
				Title:       "Settings",
				Description: "An error occurred. Settings have not been updated.",
			},
		}
	}
	respondErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: content,
		},
	})
	if respondErr != nil {
		logs.LogError(" CMND", "error responding to interaction",
			"cmd", i.ApplicationCommandData().Name,
			"err", respondErr)
	}
}
