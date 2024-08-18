package streams

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/config"
	"gamestreams/db"
	"gamestreams/logs"
	"gamestreams/utils"
)

// StreamList populates a Streams struct with upcoming streams from the streams table
// of the database. It then creates a discordgo.MessageEmbed struct with the date, time
// and title of the next [limit] streams. The limit is set in the config.toml file.
func StreamList() (*discordgo.MessageEmbed, error) {
	embed := &discordgo.MessageEmbed{
		Title: "Upcoming Streams",
		Color: config.Values.Discord.EmbedColor,
	}
	var streamList db.Streams
	if upcomErr := streamList.GetUpcoming(); upcomErr != nil {
		return nil, upcomErr
	}

	if len(streamList.Streams) == 0 {
		return embed, errors.New("no streams found")
	}
	for i, stream := range streamList.Streams {
		logs.LogInfo("STRMS", "creating embed field", false,
			"name", stream.Name,
			"time", stream.Time)

		embedField, embedErr := streamEmbedField(stream)
		if embedErr != nil {
			return nil, embedErr
		}
		//  remove the newline from the first stream embed
		if i == 0 {
			embedField.Name = strings.Split(embedField.Name, "\n")[1]
		}
		embed.Fields = append(embed.Fields, embedField)
	}
	return embed, nil
}

// StreamInfo gets a stream from the streams table of the database by name. It then
// returns a discordgo.MessageEmbed struct with the date, time, platforms, URL, and
// description of the stream.
func StreamInfo(streamName string) (*discordgo.MessageEmbed, error) {
	var streams db.Streams
	if err := streams.GetInfo(streamName); err != nil {
		return nil, err
	}
	if len(streams.Streams) == 0 {
		return nil, errors.New("no streams found")
	}
	stream := streams.Streams[0]
	date, time, dtErr := utils.CreateTimestamp(stream.Date, stream.Time)
	if dtErr != nil {
		return nil, dtErr
	}
	embed := &discordgo.MessageEmbed{
		Title: stream.Name,
		Color: config.Values.Discord.EmbedColor,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Platforms",
				Value:  stream.Platform,
				Inline: false,
			},
			{
				Name:   "\u200b\nDate",
				Value:  date,
				Inline: true,
			},
			{
				Name:   "\u200b\nTime",
				Value:  time,
				Inline: true,
			},
			{
				Name:   "\u200b\nURL",
				Value:  stream.URL,
				Inline: false,
			},
			{
				Name:   "\u200b\nDescription",
				Value:  stream.Description,
				Inline: false,
			},
		},
	}
	return embed, nil
}

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
		embed, embedErr := createStreamEmbed(stream, settings.AnnounceRole.Value)
		if embedErr != nil {
			logs.LogError("STRMS", "error creating embed",
				"server", server,
				"err", embedErr)
			continue
		}
		msg, postErr := session.ChannelMessageSendEmbed(settings.AnnounceChannel.Value, embed)
		if postErr != nil {
			logs.LogError("STRMS", "error posting message",
				"server", server,
				"channel", settings.AnnounceChannel,
				"role", settings.AnnounceRole,
				"err", postErr)
		}
		go EditAnnouncementEmbed(msg, embed, session)
	}
}

// MakeStreamURLDirect checks if the stream URL is a Youtube link and if so, gets the
// direct URL to the stream. This is done as streams could be linked to as a profile's
// /live URL which would no longer link to the correct video after the stream has ended.
func MakeStreamURLDirect(stream *db.Stream) {
	if strings.Contains(stream.URL, "youtube") {
		directUrl, success := utils.GetYoutubeDirectURL(stream.URL)
		if success {
			stream.URL = directUrl
		}
	}
}

// EditAnnouncementEmbed edits the description of an announcement embed to show that
// the stream has started. It does this by changing the "starting" to "started" in the
// description. This is achieved by creating a new goroutine that sleeps until the
// stream start time, then edits the message.
func EditAnnouncementEmbed(msg *discordgo.Message, embed *discordgo.MessageEmbed, session *discordgo.Session) {
	embed.Description = embed.Description[0:14] + "ed" + embed.Description[17:]
	medit := discordgo.NewMessageEdit(msg.ChannelID, msg.ID).SetEmbed(embed)
	time.Sleep(time.Duration(config.Values.Schedule.NotificationTMinus))
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
func createStreamEmbed(stream db.Stream, role string) (*discordgo.MessageEmbed, error) {
	ts, tsErr := utils.CreateTimestampRelative(stream.Date, stream.Time)
	if tsErr != nil {
		return nil, tsErr
	}
	if role != "" {
		role = fmt.Sprintf("<@&%s> ", role)
	}
	embed :=
		&discordgo.MessageEmbed{
			Title:       stream.Name,
			URL:         stream.URL,
			Type:        "video",
			Description: fmt.Sprintf("%s**Stream starting %s.**\n\n%s", role, ts, stream.Description),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: utils.GetVideoThumbnail(stream.URL),
			},
			Color: config.Values.Discord.EmbedColor,
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
		logs.LogInfo("STRMS", "found servers for platform", false,
			"platform", platform)

		allServerPlatforms = append(allServerPlatforms, server_list...)
	}
	return allServerPlatforms, nil
}

// streamEmbedField returns a discordgo.MessageEmbedField struct with the date, time,
// and name of the given stream.
func streamEmbedField(stream db.Stream) (*discordgo.MessageEmbedField, error) {
	ds, ts, tsErr := utils.CreateTimestamp(stream.Date, stream.Time)
	if tsErr != nil {
		return nil, tsErr
	}
	field := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("\u200b\n%s\t%s", ds, ts),
		Value:  stream.Name,
		Inline: false,
	}
	return field, nil
}
