/*
settings.go provides the /settings command. The settings command allows server owners to
update the stream announcement settings for the server.
*/
package commands

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"gamestreams/config"
	"gamestreams/db"
	"gamestreams/discord"
	"gamestreams/logs"
	"gamestreams/utils"
)

// settings updates the bot settings for the server. It parses the options from the
// interaction into an options struct. If the options are empty, it responds with the
// current settings. If the reset option is true, it resets the settings to default.
// If the options are not empty, it first gets the current settings from the database,
// then merges the new settings with the current settings into a single struct. It then
// writes the updated settings to the database and responds with the updated settings.
// If an error occurs, it responds with an error message.
func settings(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if userIsBlacklisted(i) {
		return
	}
	a := db.CommandData{}
	a.Start(i)
	defer a.End()

	userID := discord.GetUserID(i)
	logs.LogInfo(" CMND", "settings command", false,
		"user", userID,
		"server", i.GuildID)

	options := parseOptions(i.ApplicationCommandData().Options)

	var status string
	if *options == (db.Settings{}) || options.Reset {
		status = "Current settings:"
	} else {
		status = "Settings successfully updated.\n\n**Current settings:**"
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

	content := []*discordgo.MessageEmbed{
		{
			Title:       "Settings",
			Description: status,
			Color:       config.Values.Discord.EmbedColour,
			Fields: []*discordgo.MessageEmbedField{
				{},
				{
					Name:   "Announce Channel",
					Value:  utils.PlaceholderText(fmt.Sprintf("<#%s>", currentOptions.AnnounceChannel.Value)),
					Inline: false,
				},
				{
					Name:   "Announce Role",
					Value:  utils.PlaceholderText(discord.DisplayRole(s, i.GuildID, currentOptions.AnnounceRole.Value)),
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
			{
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

// parseOptions parses the options from the interaction into a settings struct.
func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) *db.Settings {
	var s db.Settings
	for _, option := range options {
		switch option.Name {
		case "channel":
			s.AnnounceChannel.Value = option.Value.(string)
			s.AnnounceChannel.Set = true
		case "role":
			s.AnnounceRole.Value = option.Value.(string)
			s.AnnounceRole.Set = true
		case "playstation":
			s.Playstation.Value = option.BoolValue()
			s.Playstation.Set = true
		case "xbox":
			s.Xbox.Value = option.BoolValue()
			s.Xbox.Set = true
		case "nintendo":
			s.Nintendo.Value = option.BoolValue()
			s.Nintendo.Set = true
		case "pc":
			s.PC.Value = option.BoolValue()
			s.PC.Set = true
		case "vr":
			s.VR.Value = option.BoolValue()
			s.VR.Set = true
		case "reset":
			s.Reset = option.BoolValue()
		}
	}
	return &s
}
