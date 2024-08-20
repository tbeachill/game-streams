/*
announcements.go contains functions for scheduling and posting stream announcements.
*/
package streams

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/config"
	"gamestreams/db"
	"gamestreams/logs"
	"gamestreams/utils"
)

// ScheduleNotifications gets all streams for today that have not yet started from the
// streams table of the database. It then schedules notifications for each stream by
// creating a goroutine for each stream that sleeps until the streams start time -
// the configured notification_t_minus in config.toml. It then posts a message to the
// servers that are following one or more of the platforms of the stream by calling the
// PostStreamLink function.
func ScheduleNotifications(session *discordgo.Session) error {
	var streamList db.Streams
	if todayErr := streamList.GetToday(); todayErr != nil {
		return todayErr
	}
	if len(streamList.Streams) == 0 {
		logs.LogInfo("STRMS", "no streams today", false)
		return nil
	}
	for i, stream := range streamList.Streams {
		go func(currentStream *db.Stream) {
			dateTime := fmt.Sprintf("%s %s", currentStream.Date, currentStream.Time)
			streamTime, parseErr := time.Parse("2006-01-02 15:04", dateTime)
			if parseErr != nil {
				logs.LogError("STRMS", "error parsing time",
					"err", parseErr)
				return
			}
			minsBefore := time.Minute * time.Duration(config.Values.Schedule.NotificationTMinus)
			timeToStream := streamTime.Sub(time.Now().UTC()) - minsBefore
			time.Sleep(timeToStream)
			PostStreamLink(*currentStream, session)
		}(&stream)
		logs.LogInfo("STRMS", "scheduled stream", false,
			"goroutine", i+1,
			"name", stream.Name,
			"time", stream.Time)
	}
	streamLen := len(streamList.Streams)
	logs.LogInfo("STRMS", "scheduled todays streams", false,
		"count", streamLen)
	return nil
}

// PostStreamLink posts an embed with the given streams information to the servers
// that are following one or more of the platforms of the stream and has an announcement
// channel set.
func PostStreamLink(stream db.Stream, session *discordgo.Session) {
	logs.LogInfo("STRMS", "posting stream link", false,
		"stream", stream.Name,
		"platforms", stream.Platform)

	allServerPlatforms, platErr := getAllPlatforms(stream)
	if platErr != nil {
		logs.LogError("STRMS", "error getting server platforms",
			"err", platErr)
		return
	}
	// Removing duplicates is necessary because a server may follow multiple platforms
	// and the stream may be related to multiple platforms. Therefore the same server
	// may be added to allServerPlatforms multiple times.
	uniqueServers := utils.RemoveSliceDuplicates(allServerPlatforms)
	MakeStreamURLDirect(&stream)

	logs.LogInfo("STRMS", "retrieved server IDs", false,
		"count", len(uniqueServers))

	for server := range uniqueServers {
		var settings db.Settings
		if getSetErr := settings.Get(server); getSetErr != nil {
			logs.LogError("SCHED", "error getting settings",
				"server", server,
				"err", getSetErr)
			continue
		}
		if settings.AnnounceChannel.Value == "" {
			continue
		}
		embed, embedErr := createStreamEmbed(stream)
		if embedErr != nil {
			logs.LogError("STRMS", "error creating embed",
				"server", server,
				"err", embedErr)
			continue
		}
		msg, postErr := session.ChannelMessageSendComplex(settings.AnnounceChannel.Value, &discordgo.MessageSend{
			Embed:   embed,
			Content: fmt.Sprintf("<@&%s>", settings.AnnounceRole.Value),
		})
		if postErr != nil {
			logs.LogError("STRMS", "error posting message",
				"server", server,
				"channel", settings.AnnounceChannel,
				"role", settings.AnnounceRole,
				"err", postErr)
		}
		go EditAnnouncementEmbed(msg, embed, session)
	}
	logs.LogInfo("STRMS", "finished posting stream", false,
		"stream", stream.Name)
}

// EditAnnouncementEmbed edits the description of an announcement embed to show that
// the stream has started. It does this by changing the "starting" to "started" in the
// description. This is achieved by creating a new goroutine that sleeps until the
// stream start time, then edits the message.
func EditAnnouncementEmbed(msg *discordgo.Message, embed *discordgo.MessageEmbed, session *discordgo.Session) {
	embed.Description = embed.Description[0:14] + "ed" + embed.Description[17:]
	medit := discordgo.NewMessageEdit(msg.ChannelID, msg.ID).SetEmbed(embed)
	time.Sleep(time.Duration(config.Values.Schedule.NotificationTMinus) * time.Minute)
	_, editErr := session.ChannelMessageEditComplex(medit)
	if editErr != nil {
		logs.LogError("STRMS", "error editing message",
			"channel", msg.ChannelID,
			"message", msg.ID,
			"err", editErr)
	}
}

// createStreamEmbed returns a discordgo.MessageEmbed struct with the stream
// information from the given stream and announcement role.
func createStreamEmbed(stream db.Stream) (*discordgo.MessageEmbed, error) {
	ts, tsErr := utils.CreateTimestampRelative(stream.Date, stream.Time)
	if tsErr != nil {
		return nil, tsErr
	}
	embed :=
		&discordgo.MessageEmbed{
			Title:       stream.Name,
			URL:         stream.URL,
			Type:        "video",
			Description: fmt.Sprintf("**Stream starting %s.**\n\n%s", ts, stream.Description),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: utils.GetVideoThumbnail(stream.URL),
			},
			Color: config.Values.Discord.EmbedColour,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "\u200b\nPlatforms",
					Value:  stream.Platform,
					Inline: false,
				},
			},
		}
	return embed, nil
}

// getAllPlatforms returns a slice of server IDs that are following one or more of the
// platforms of the given stream.
func getAllPlatforms(stream db.Stream) ([]string, error) {
	platforms := strings.Split(stream.Platform, ",")
	var allServerPlatforms []string
	for _, platform := range platforms {
		platform = strings.Trim(platform, " ")
		server_list, platErr := db.GetPlatformServerIDs(platform)
		if platErr != nil {
			return nil, platErr
		}
		allServerPlatforms = append(allServerPlatforms, server_list...)
	}
	return allServerPlatforms, nil
}
