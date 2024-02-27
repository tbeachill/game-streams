package bot

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreambot/commands"
	"gamestreambot/db"
	"gamestreambot/streams"
)

// TODO: add log when added to new server?
// TODO: how to see how many servers its in?
// TODO: add uptime command
// TODO: improve logging

func Run(botToken, appID string) {
	session, sessionErr := discordgo.New("Bot " + botToken)
	if sessionErr != nil {
		log.Println("error creating Discord session: ", sessionErr)
		return
	}
	if openErr := session.Open(); openErr != nil {
		log.Println("error connecting to Discord: ", openErr)
		return
	}
	defer session.Close()

	commands.RegisterCommands(appID, session)
	commands.RegisterHandler(session, &discordgo.InteractionCreate{})
	go startUpdater()
	go startScheduler(session)

	log.Println("running: press ctrl + c to terminate")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

// check for updates to the streams every hour, on the hour
func startUpdater() {
UPDATE:
	log.Println("checking for stream updates...")
	if updateErr := db.UpdateStreams(); updateErr != nil {
		log.Println("error updating streams: ", updateErr)
	}
	for {
		time.Sleep(1 * time.Minute)
		if time.Now().Minute() == 0 {
			goto UPDATE
		}
	}
}

// check if a new day has started, if so, schedule notifications for today's streams
func startScheduler(session *discordgo.Session) {
SCHEDULE:
	log.Println("scheduling notifications for today's streams...")
	if scheduleErr := streams.ScheduleNotifications(session); scheduleErr != nil {
		log.Println("error scheduling today's streams: ", scheduleErr)
	}
	for {
		time.Sleep(1 * time.Minute)
		hour, min, _ := time.Now().UTC().Clock()
		if hour == 0 && min == 0 {
			goto SCHEDULE
		}
	}
}
