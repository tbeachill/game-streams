package streams

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreambot/db"
	"gamestreambot/utils"
)

// return a list of all upcoming streams as a slice of embeds
func StreamList() (*discordgo.MessageEmbed, error) {
	embed := &discordgo.MessageEmbed{
		Title: "Upcoming Streams",
		Color: 0xc3d23e,
	}
	var streamList db.Streams
	streamList.GetUpcoming()

	if len(streamList.Streams) == 0 {
		return embed, errors.New("no streams found")
	}
	for _, stream := range streamList.Streams {
		utils.Log.Info.WithPrefix(" CMND").Info("creating embed field", "name", stream.Name, "time", stream.Time)
		embed.Fields = append(embed.Fields, streamEmbedField(stream))
	}
	return embed, nil
}

// create a goroutine to sleep until 5 minutes before the stream, then run the notification function
func ScheduleNotifications(session *discordgo.Session) error {
	var streamList db.Streams
	streamList.GetToday()

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
				return
			}
			timeToStream := streamTime.Sub(time.Now().UTC()) - (time.Minute * 5)
			time.Sleep(timeToStream)
			PostStreamLink(*currentStream, session)
		}(&stream)
		utils.Log.Info.WithPrefix("SCHED").Info("scheduled stream", "goroutine", i+1, "name", stream.Name, "time", stream.Time)
	}
	streamLen := len(streamList.Streams)
	utils.Log.Info.WithPrefix("SCHED").Infof("scheduled %d stream%s for today", streamLen, utils.Pluralise(streamLen))
	return nil
}

// post a message to every server that is following the platform when a stream is about to start
func PostStreamLink(stream db.Stream, session *discordgo.Session) {
	utils.Log.Info.WithPrefix("SCHED").Info("posting stream link to subscribed servers", "stream", stream.Name, "platforms", stream.Platform)
	allServerPlatforms := getAllPlatforms(stream)
	uniqueServers := utils.RemoveSliceDuplicates(allServerPlatforms)

	for server := range uniqueServers {
		var options db.Options
		options.Get(server)
		_, postErr := session.ChannelMessageSendEmbed(options.AnnounceChannel.Value, createStreamEmbed(stream, options.AnnounceRole.Value))
		if postErr != nil {
			utils.Log.ErrorWarn.WithPrefix("SCHED").Error("error posting message", "server", server, "channel", options.AnnounceChannel, "role", options.AnnounceRole, "err", postErr)
		}
	}
}

// create an embed for a stream
func createStreamEmbed(stream db.Stream, role string) *discordgo.MessageEmbed {
	ts, tsErr := utils.CreateTimestampRelative(stream.Date, stream.Time)
	if tsErr != nil {
		utils.Log.ErrorWarn.WithPrefix("SCHED").Error("error creating timestamp")
		return nil
	}
	if role != "" {
		role = fmt.Sprintf("<@&%s> ", role)
	}
	embed :=
		&discordgo.MessageEmbed{
			Title:       stream.Name,
			URL:         stream.URL,
			Type:        "video",
			Description: fmt.Sprintf("%sstream starting %s\n\n%s", role, ts, stream.Description),
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
	return embed
}

func getAllPlatforms(stream db.Stream) []string {
	platforms := strings.Split(stream.Platform, ",")
	var allServerPlatforms []string
	for _, platform := range platforms {
		platform = strings.Trim(platform, " ")
		server_list, platErr := db.GetPlatformServerIDs(platform)
		if platErr != nil {
			utils.Log.ErrorWarn.WithPrefix("SCHED").Error("error getting platform server IDs")
			return nil
		}
		utils.Log.Info.WithPrefix("SCHED").Info("found servers for platform", "platform", platform)
		allServerPlatforms = append(allServerPlatforms, server_list...)
	}
	return allServerPlatforms
}

// return an embed field with the stream information
func streamEmbedField(stream db.Stream) *discordgo.MessageEmbedField {
	ds, ts, tsErr := utils.CreateTimestamp(stream.Date, stream.Time)
	if tsErr != nil {
		utils.Log.ErrorWarn.WithPrefix(" CMND").Error("error creating timestamp")
		return nil
	}
	field := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s %s", ds, ts),
		Value:  fmt.Sprintf("**%s**", stream.Name),
		Inline: false,
	}
	return field
}
