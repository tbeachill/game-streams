package streams

// TODO: limit output to number of days in the future
// TODO: write tests

import (
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"

	"gamestreambot/db"
	"gamestreambot/utils"
)

// return a string of all streams for display on discord
func StreamList() (string, error) {
	streamList, dbErr := db.GetUpcomingStreams()
	if dbErr != nil {
		utils.EWLogger.WithPrefix(" MAIN").Error("error getting streams from db")
		return "", dbErr
	}
	if len(streamList.Streams) == 0 {
		return "### No upcoming streams.", nil
	}
	message := "### Upcoming Streams:\n"

	for i, stream := range streamList.Streams {
		ts, tsErr := utils.CreateTimestamp(stream.Date, stream.Time)
		if tsErr != nil {
			utils.EWLogger.WithPrefix(" MAIN").Error("error creating timestamp")
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
	for _, stream := range streamList.Streams {
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
	}
	streamLen := len(streamList.Streams)
	utils.Logger.WithPrefix("SCHED").Infof("scheduled %d stream%s for today", streamLen, utils.Pluralise(streamLen))
	return nil
}

// post a message to the server when a stream is about to start
func postStreamLink(stream db.Stream, session *discordgo.Session) {
	// TODO: rewrite this function to work with all servers
	godotenv.Load()
	channelID := os.Getenv("CHANNEL_ID")
	session.ChannelMessageSend(channelID, fmt.Sprintf("Stream starting in 5 minutes: %s", stream.URL))
}
