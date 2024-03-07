package bot

import (
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreambot/commands"
	"gamestreambot/db"
	"gamestreambot/stats"
	"gamestreambot/streams"
	"gamestreambot/utils"
)

// TODO: add help command
// TODO: look at structs and turn some functions into methods
// TODO: add uptime command
// TODO: check error handling in all functions
// TODO: check logging in all functions
// TODO: set up test bot

func Run(botToken, appID string) {
	session, sessionErr := discordgo.New("Bot " + botToken)
	if sessionErr != nil {
		utils.EWLogger.WithPrefix(" MAIN").Error("error creating Discord session", "err", sessionErr)
		return
	}
	if openErr := session.Open(); openErr != nil {
		utils.EWLogger.WithPrefix(" MAIN").Error("error connecting to Discord", "err", openErr)
		return
	}
	defer session.Close()
	//commands.RemoveAllCommands(appID, session)
	commands.RegisterCommands(appID, session)
	commands.RegisterHandler(session, &discordgo.InteractionCreate{})
	go startUpdater()
	go startScheduler(session)
	stats.MonitorGuilds(session)

	utils.Logger.WithPrefix(" MAIN").Info("running. press ctrl + c to terminate")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

// TODO: change to sleep until time to run updater
// check for updates to the streams every hour, on the hour
func startUpdater() {
UPDATE:
	utils.Logger.WithPrefix("UPDAT").Info("checking for stream updates...")
	if updateErr := db.UpdateStreams(); updateErr != nil {
		utils.EWLogger.WithPrefix("UPDAT").Error("error updating streams", "err", updateErr)
	}
	for {
		time.Sleep(1 * time.Minute)
		if time.Now().Minute() == 0 {
			goto UPDATE
		}
	}
}

// TODO: change to sleep until time to run scheduler
// check if a new day has started, if so, schedule notifications for today's streams
func startScheduler(session *discordgo.Session) {
SCHEDULE:
	utils.Logger.WithPrefix("SCHED").Info("scheduling notifications for today's streams...")
	if scheduleErr := streams.ScheduleNotifications(session); scheduleErr != nil {
		utils.EWLogger.WithPrefix("SCHED").Error("error scheduling today's streams", "err", scheduleErr)
	}
	for {
		time.Sleep(1 * time.Minute)
		hour, min, _ := time.Now().UTC().Clock()
		if hour == 0 && min == 0 {
			goto SCHEDULE
		}
	}
}
