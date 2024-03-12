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

// TODO: document functions properly - arguments, return values
// TODO: look at structs and turn some functions into methods
// TODO: check error handling in all functions - sql should have end of function error checks
// TODO: check logging in all functions
// TODO: message me if there is an error
// TODO: message me weekly stats

func Run(botToken, appID string) {
	session, sessionErr := discordgo.New("Bot " + botToken)
	if sessionErr != nil {
		utils.Log.ErrorWarn.WithPrefix(" MAIN").Error("error creating Discord session", "err", sessionErr)
		return
	}
	if openErr := session.Open(); openErr != nil {
		utils.Log.ErrorWarn.WithPrefix(" MAIN").Error("error connecting to Discord", "err", openErr)
		return
	}
	defer session.Close()
	//commands.RemoveAllCommands(appID, session)
	commands.RegisterCommands(appID, session)
	commands.RegisterHandler(session, &discordgo.InteractionCreate{})
	go startUpdater()
	go startScheduler(session)
	stats.MonitorGuilds(session)

	utils.Log.Info.WithPrefix(" MAIN").Info("running. press ctrl + c to terminate")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

// check for updates to the streams every hour, on the hour
func startUpdater() {
	for {
		utils.Log.Info.WithPrefix("UPDAT").Info("checking for stream updates...")
		if updateErr := db.UpdateStreams(); updateErr != nil {
			utils.Log.ErrorWarn.WithPrefix("UPDAT").Error("error updating streams", "err", updateErr)
		}
		minsRemaining := 60 - time.Now().UTC().Minute()
		utils.Log.Info.WithPrefix("UPDAT").Info("sleeping until next update", "minutes", minsRemaining)
		time.Sleep(time.Duration(minsRemaining) * time.Minute)
	}
}

// schedule notifications for today's streams every day at midnight UTC
func startScheduler(session *discordgo.Session) {
	startTime := time.Now().UTC()
	for {
		utils.Log.Info.WithPrefix("SCHED").Info("scheduling notifications for today's streams...")
		if scheduleErr := streams.ScheduleNotifications(session); scheduleErr != nil {
			utils.Log.ErrorWarn.WithPrefix("SCHED").Error("error scheduling today's streams", "err", scheduleErr)
		}
		hour, min, _ := time.Now().UTC().Clock()
		hoursRemaining := 24 - hour
		minsRemaining := 60 - min
		utils.Log.Info.WithPrefix("SCHED").Info("sleeping until next day", "hours", hoursRemaining, "minutes", minsRemaining)
		time.Sleep(time.Duration(hoursRemaining*60+minsRemaining) * time.Minute)
		utils.Log.Info.WithPrefix("SCHED").Info("we survived another day", "uptime", time.Now().UTC().Sub(startTime))
	}
}
