package bot

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreambot/commands"
	"gamestreambot/db"
	"gamestreambot/reports"
	"gamestreambot/stats"
	"gamestreambot/streams"
	"gamestreambot/utils"
)

// TODO: document functions properly - arguments, return values
// TODO: Add function to alert me when a stream with no time is coming up
// TODO: Add description command to return the description of a stream and the platforms

func Run(botToken, appID string) {
	session, sessionErr := discordgo.New("Bot " + botToken)
	if sessionErr != nil {
		utils.Log.ErrorWarn.WithPrefix(" MAIN").Error("error creating Discord session", "err", sessionErr)
		reports.DM(session, fmt.Sprintf("error creating Discord session:\n\terr=%s", sessionErr))
		return
	}
	if openErr := session.Open(); openErr != nil {
		utils.Log.ErrorWarn.WithPrefix(" MAIN").Error("error connecting to Discord", "err", openErr)
		reports.DM(session, fmt.Sprintf("error connecting to Discord:\n\terr=%s", openErr))
		return
	}
	defer session.Close()
	utils.RegisterSession(session)
	//commands.RemoveAllCommands(appID, session)
	commands.RegisterCommands(appID, session)
	commands.RegisterHandler(session, &discordgo.InteractionCreate{})
	go startUpdater()
	go startScheduler(session)
	stats.MonitorGuilds(session)
	reports.DM(session, "bot started")
	utils.Log.Info.WithPrefix(" MAIN").Info("running. press ctrl + c to terminate")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

// check for updates to the streams every hour, on the hour
func startUpdater() {
	var s db.Streams
	for {
		utils.Log.Info.WithPrefix("UPDAT").Info("checking for stream updates...")
		if updateErr := s.Update(); updateErr != nil {
			utils.Log.ErrorWarn.WithPrefix("UPDAT").Error("error updating streams", "err", updateErr)
			reports.DM(utils.Session, fmt.Sprintf("error updating streams:\n\terr=%s", updateErr))
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
			reports.DM(utils.Session, fmt.Sprintf("error scheduling todays streams:\n\terr=%s", scheduleErr))
		}
		hour, min, _ := time.Now().UTC().Clock()
		hoursRemaining := 24 - hour
		minsRemaining := 60 - min
		utils.Log.Info.WithPrefix("SCHED").Info("sleeping until next day", "hours", hoursRemaining, "minutes", minsRemaining)
		time.Sleep(time.Duration(hoursRemaining*60+minsRemaining) * time.Minute)
		utils.Log.Info.WithPrefix("SCHED").Info("we survived another day", "uptime", time.Now().UTC().Sub(startTime))
		reports.DM(utils.Session, fmt.Sprintf("we survived another day\n\tuptime=%s\n\tservers=%d", time.Now().UTC().Sub(startTime), stats.GetGuildNumber(session)))
	}
}
