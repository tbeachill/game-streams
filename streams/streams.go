package streams

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreambot/db"
	"gamestreambot/utils"
)

// return a string of all streams for display on discord
func StreamList() (string, error) {
	utils.Logger.WithPrefix(" CMND").Info("getting upcoming streams")
	streamList, dbErr := db.GetUpcomingStreams()
	if dbErr != nil {
		utils.EWLogger.WithPrefix(" CMND").Error("error getting streams from db")
		return "", dbErr
	}
	if len(streamList.Streams) == 0 {
		return "### No upcoming streams.", nil
	}
	message := "### Upcoming Streams:\n"

	for i, stream := range streamList.Streams {
		ts, tsErr := utils.CreateTimestamp(stream.Date, stream.Time)
		if tsErr != nil {
			utils.EWLogger.WithPrefix(" CMND").Error("error creating timestamp")
			return "", tsErr
		}
		// add date header if it's the first stream or the date is different from the previous stream
		if i == 0 || streamList.Streams[i-1].Date != stream.Date {
			message += fmt.Sprintf("\n<t:%s:d>:\n", ts)
		}
		// add stream name and timestamp
		message += fmt.Sprintf("<t:%s:t> %s\n", ts, stream.Name)
	}
	return message, nil
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
