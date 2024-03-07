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
	var embed *discordgo.MessageEmbed
	embed = &discordgo.MessageEmbed{
		Title: "Upcoming Streams",
		Color: 0xc3d23e,
	}
	streamList, dbErr := db.GetUpcomingStreams()
	if dbErr != nil {
		utils.EWLogger.WithPrefix(" CMND").Error("error getting streams from db")
		return embed, dbErr
	}
	if len(streamList.Streams) == 0 {
		return embed, errors.New("no streams found")
	}
	for _, stream := range streamList.Streams {
		utils.Logger.WithPrefix(" CMND").Info("creating embed field", "name", stream.Name, "time", stream.Time)
		embed.Fields = append(embed.Fields, streamEmbedField(stream))
	}
	return embed, nil
}

// create a goroutine to sleep until 5 minutes before the stream, then run the notification function
func ScheduleNotifications(session *discordgo.Session) error {
	streamList, dbErr := db.GetTodaysStreams()
	if dbErr != nil {
		utils.EWLogger.WithPrefix("SCHED").Error("error getting streams from db")
		return dbErr
	}
	if len(streamList.Streams) == 0 {
		utils.Logger.WithPrefix("SCHED").Info("no streams today")
		return nil
	}
	for i, stream := range streamList.Streams {
		go func(currentStream *db.Stream) {
			dateTime := fmt.Sprintf("%s %s", currentStream.Date, currentStream.Time)
			streamTime, parseErr := time.Parse("2006-01-02 15:04", dateTime)
			if parseErr != nil {
				utils.EWLogger.WithPrefix("SCHED").Error("error parsing time")
				return
			}
			timeToStream := streamTime.Sub(time.Now().UTC()) - (time.Minute * 5)
			time.Sleep(timeToStream)
			postStreamLink(*currentStream, session)
		}(&stream)
		utils.Logger.WithPrefix("SCHED").Info("scheduled stream", "goroutine", i+1, "name", stream.Name, "time", stream.Time)
	}
	streamLen := len(streamList.Streams)
	utils.Logger.WithPrefix("SCHED").Infof("scheduled %d stream%s for today", streamLen, utils.Pluralise(streamLen))
	return nil
}

// post a message to every server that is following the platform when a stream is about to start
func postStreamLink(stream db.Stream, session *discordgo.Session) {
	allServerPlatforms := getAllPlatforms(stream)
	allServerPlatforms = utils.RemoveSliceDuplicates(allServerPlatforms)
	utils.Logger.WithPrefix("SCHED").Info("posting stream link to subscribed servers", "stream", stream.Name)

	for _, server := range allServerPlatforms {
		options, getErr := db.GetOptions(server)
		if getErr != nil {
			utils.EWLogger.WithPrefix("SCHED").Error("error getting server options", "server", server, "err", getErr)
			continue
		}
		channel, channelErr := session.State.Channel(options.AnnounceChannel)
		if channelErr != nil {
			utils.EWLogger.WithPrefix("SCHED").Error("error getting channel", "server", server, "channel", options.AnnounceChannel, "err", channelErr)
			continue
		}
		if options.AnnounceRole != "" {
			_, postErr := session.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
				Content: fmt.Sprintf("<@&%s> Stream starting in 5 minutes: %s", options.AnnounceRole, stream.URL),
			})
			if postErr != nil {
				utils.EWLogger.WithPrefix("SCHED").Error("error posting message", "server", server, "channel", channel.ID, "role", options.AnnounceRole, "err", postErr)
			}
			continue
		}
		_, postErr := session.ChannelMessageSend(channel.ID, fmt.Sprintf("Stream starting in 5 minutes: %s", stream.URL))
		if postErr != nil {
			utils.EWLogger.WithPrefix("SCHED").Error("error posting message", "server", server, "channel", channel.ID, "err", postErr)
		}
	}
}

func getAllPlatforms(stream db.Stream) []string {
	platforms := strings.Split(stream.Platform, ",")
	var allServerPlatforms []string
	for _, platform := range platforms {
		server_list, platErr := db.GetPlatformServerIDs(platform)
		if platErr != nil {
			utils.EWLogger.WithPrefix("SCHED").Error("error getting platform server IDs")
			return nil
		}
		allServerPlatforms = append(allServerPlatforms, server_list...)
	}
	return allServerPlatforms
}

// return an embed field with the stream information
func streamEmbedField(stream db.Stream) *discordgo.MessageEmbedField {
	ds, ts, tsErr := utils.CreateTimestamp(stream.Date, stream.Time)
	if tsErr != nil {
		utils.EWLogger.WithPrefix(" CMND").Error("error creating timestamp")
		return nil
	}
	field := &discordgo.MessageEmbedField{
		Value:  fmt.Sprintf("%s%s %s", ds, ts, stream.Name),
		Inline: false,
	}
	return field
}
