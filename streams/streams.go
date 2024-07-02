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

// return a list of all upcoming streams as a slice of embeds
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

// get the information for a stream
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

// create a goroutine to sleep until 10 minutes before the stream, then run the notification function
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
			timeToStream := streamTime.Sub(time.Now().UTC()) - (time.Minute * 10)
			time.Sleep(timeToStream)
			PostStreamLink(*currentStream, session)
		}(&stream)
		utils.Log.Info.WithPrefix("SCHED").Info("scheduled stream", "goroutine", i+1, "name", stream.Name, "time", stream.Time)
	}
	streamLen := len(streamList.Streams)
	utils.Log.Info.WithPrefix("SCHED").Infof("scheduled %d stream%s for today", streamLen, utils.Pluralise(streamLen))
	return nil
}

// post a message to every server that is following the platform  and has a channel set when a stream is about to start
func PostStreamLink(stream db.Stream, session *discordgo.Session) {
	utils.Log.Info.WithPrefix("SCHED").Info("posting stream link to subscribed servers", "stream", stream.Name, "platforms", stream.Platform)
	allServerPlatforms, platErr := getAllPlatforms(stream)
	if platErr != nil {
		utils.Log.ErrorWarn.WithPrefix("SCHED").Error("error getting server platforms", "err", platErr)
		reports.DM(session, fmt.Sprintf("error getting server platforms:\n\terr=%s", platErr))
		return
	}
	uniqueServers := utils.RemoveSliceDuplicates(allServerPlatforms)

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
		_, postErr := session.ChannelMessageSendEmbed(options.AnnounceChannel.Value, embed)
		if postErr != nil {
			utils.Log.ErrorWarn.WithPrefix("SCHED").Error("error posting message", "server", server, "channel", options.AnnounceChannel, "role", options.AnnounceRole, "err", postErr)
			reports.DM(session, fmt.Sprintf("error posting message:\n\tserver=%s\n\tchannel=%s\n\trole=%s\n\terr=%s", server, options.AnnounceChannel.Value, options.AnnounceRole.Value, postErr))
		}
	}
}

// create an embed for a stream
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
			Description: fmt.Sprintf("%sStream starting %s\n\n%s", role, ts, stream.Description),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: utils.GetVideoThumbnail(stream.URL),
			},
			Color: 0xc3d23e,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Platforms",
					Value:  stream.Platform,
					Inline: false,
				},
			},
		}
	return embed, nil
}

// split the list of platforms then search the database for servers following each platforms
// return a slice of server IDs
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

// return an embed field with the stream information
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
