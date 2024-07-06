package streams

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreambot/db"
	"gamestreambot/reports"
	"gamestreambot/utils"
)

// STREAM_T_MINUS is the time before a stream to send a notification
var STREAM_T_MINUS = time.Minute * 10

// StreamList populates a Streams struct with upcoming streams from the streams table of the database.
// It then creates a discordgo.MessageEmbed struct with the date, time and title of the next 10 upcoming streams.
func StreamList() (*discordgo.MessageEmbed, error) {
	embed := &discordgo.MessageEmbed{
		Title: "Upcoming Streams",
		Color: 0xc3d23e,
	}
	var streamList db.Streams
	if upcomErr := streamList.GetUpcoming(); upcomErr != nil {
		return nil, upcomErr
	}

	if len(streamList.Streams) == 0 {
		return embed, errors.New("no streams found")
	}
	for i, stream := range streamList.Streams {
		utils.Log.Info.WithPrefix(" CMND").Info("creating embed field", "name", stream.Name, "time", stream.Time)
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

// StreamInfo gets a stream from the streams table of the database by name. It then creates a discordgo.MessageEmbed
// struct with the date, time, platforms, URL, and description of the stream.
func StreamInfo(streamName string) (*discordgo.MessageEmbed, error) {
	utils.Log.Info.WithPrefix(" CMND").Info("getting stream info", "name", streamName)
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
		Color: 0xc3d23e,
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

// ScheduleNotifications gets all streams for today that have not yet started from the streams table of the database.
// It then schedules notifications for each stream by creating a goroutine for each stream that sleeps until the
// streams start time - STREAM_T_MINUS, then posts a message to the servers that are following one or more of the
// platforms of the stream by calling the PostStreamLink function.
func ScheduleNotifications(session *discordgo.Session) error {
	var streamList db.Streams
	if todayErr := streamList.GetToday(); todayErr != nil {
		return todayErr
	}
	if len(streamList.Streams) == 0 {
		utils.Log.Info.WithPrefix("SCHED").Info("no streams today")
		return nil
	}
	for i, stream := range streamList.Streams {
		go func(currentStream *db.Stream) {
			dateTime := fmt.Sprintf("%s %s", currentStream.Date, currentStream.Time)
			streamTime, parseErr := time.Parse("2006-01-02 15:04", dateTime)
			if parseErr != nil {
				utils.Log.ErrorWarn.WithPrefix("SCHED").Error("error parsing time")
				reports.DM(session, fmt.Sprintf("error parsing time:\n\terr=%s", parseErr))
				return
			}
			timeToStream := streamTime.Sub(time.Now().UTC()) - STREAM_T_MINUS
			time.Sleep(timeToStream)
			PostStreamLink(*currentStream, session)
		}(&stream)
		utils.Log.Info.WithPrefix("SCHED").Info("scheduled stream", "goroutine", i+1, "name", stream.Name, "time", stream.Time)
	}
	streamLen := len(streamList.Streams)
	utils.Log.Info.WithPrefix("SCHED").Infof("scheduled %d stream%s for today", streamLen, utils.Pluralise(streamLen))
	return nil
}

// PostStreamLink posts an embed with the given streams information to the servers that are following one or more of
// the platforms of the stream and has an announcement channel set.
func PostStreamLink(stream db.Stream, session *discordgo.Session) {
	utils.Log.Info.WithPrefix("SCHED").Info("posting stream link to subscribed servers", "stream", stream.Name, "platforms", stream.Platform)
	allServerPlatforms, platErr := getAllPlatforms(stream)
	if platErr != nil {
		utils.Log.ErrorWarn.WithPrefix("SCHED").Error("error getting server platforms", "err", platErr)
		reports.DM(session, fmt.Sprintf("error getting server platforms:\n\terr=%s", platErr))
		return
	}
	// Removing duplicates is necessary because a server may follow multiple platforms and the stream may be
	// related to multiple platforms. Therefore the same server may be added to allServerPlatforms multiple times.
	uniqueServers := utils.RemoveSliceDuplicates(allServerPlatforms)
	MakeStreamUrlDirect(&stream)

	for server := range uniqueServers {
		var options db.Options
		if getOptErr := options.Get(server); getOptErr != nil {
			utils.Log.ErrorWarn.WithPrefix("SCHED").Error("error getting options", "server", server, "err", getOptErr)
			reports.DM(session, fmt.Sprintf("error getting options:\n\tserver=%s\n\terr=%s", server, getOptErr))
			continue
		}
		if options.AnnounceChannel.Value == "" {
			continue
		}
		embed, embedErr := createStreamEmbed(stream, options.AnnounceRole.Value)
		if embedErr != nil {
			utils.Log.ErrorWarn.WithPrefix("SCHED").Error("error creating embed", "server", server, "err", embedErr)
			reports.DM(session, fmt.Sprintf("error creating embed:\n\tserver=%s\n\terr=%s", server, embedErr))
			continue
		}
		msg, postErr := session.ChannelMessageSendEmbed(options.AnnounceChannel.Value, embed)
		if postErr != nil {
			utils.Log.ErrorWarn.WithPrefix("SCHED").Error("error posting message", "server", server, "channel", options.AnnounceChannel, "role", options.AnnounceRole, "err", postErr)
			reports.DM(session, fmt.Sprintf("error posting message:\n\tserver=%s\n\tchannel=%s\n\trole=%s\n\terr=%s", server, options.AnnounceChannel.Value, options.AnnounceRole.Value, postErr))
		}
		go EditAnnouncementEmbed(msg, embed, session)
	}
}

// MakeStreamUrlDirect checks if the stream URL is a Youtube link and if so, gets the direct URL to the stream.
// This is done as streams could be linked to as a profile's /live URL which would no longer link to the correct video
// after the stream has ended.
func MakeStreamUrlDirect(stream *db.Stream) {
	if strings.Contains(stream.URL, "youtube") {
		directUrl, success := utils.GetYoutubeDirectUrl(stream.URL)
		if success {
			stream.URL = directUrl
		}
	}
}

// EditAnnouncementEmbed edits the description of an announcement embed to show that the stream has started.
// It does this by changing the "starting" to "started" in the description. This is achieved by creating a new
// goroutine that sleeps for STREAM_T_MINUS then edits the message.
func EditAnnouncementEmbed(msg *discordgo.Message, embed *discordgo.MessageEmbed, session *discordgo.Session) {
	embed.Description = embed.Description[0:14] + "ed" + embed.Description[17:]
	medit := discordgo.NewMessageEdit(msg.ChannelID, msg.ID).SetEmbed(embed)
	time.Sleep(STREAM_T_MINUS)
	_, editErr := session.ChannelMessageEditComplex(medit)
	if editErr != nil {
		utils.Log.ErrorWarn.WithPrefix("SCHED").Error("error editing message", "channel", msg.ChannelID, "message", msg.ID, "err", editErr)
		reports.DM(session, fmt.Sprintf("error editing message:\n\tchannel=%s\n\tmessage=%s\n\terr=%s", msg.ChannelID, msg.ID, editErr))
	}
}

// createStreamEmbed returns a discordgo.MessageEmbed struct with the stream information from the given stream and
// announcement role.
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
			Color: 0xc3d23e,
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

// getAllPlatforms returns a slice of server IDs that are following one or more of the platforms of the given stream.
func getAllPlatforms(stream db.Stream) ([]string, error) {
	platforms := strings.Split(stream.Platform, ",")
	var allServerPlatforms []string
	for _, platform := range platforms {
		platform = strings.Trim(platform, " ")
		server_list, platErr := db.GetPlatformServerIDs(platform)
		if platErr != nil {
			return nil, platErr
		}
		utils.Log.Info.WithPrefix("SCHED").Info("found servers for platform", "platform", platform)
		allServerPlatforms = append(allServerPlatforms, server_list...)
	}
	return allServerPlatforms, nil
}

// streamEmbedField returns a discordgo.MessageEmbedField struct with the date, time, and name of the given stream.
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
