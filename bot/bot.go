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

// Run is the main function that runs the bot. It creates a new Discord session, registers the commands,
// and starts the updater and scheduler goroutines. The bot runs until it receives a termination signal (ctrl + c).
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

// startUpdater creates a new Streams struct and updates the streams every hour by running the Update method.
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

// startScheduler runs at startup, then at midnight UTC every day. It checks if there are any streams tomorrow with no
// time set and alerts me to add a time. It creates a streams struct with today's streams and schedules notifications
// for each stream by running the ScheduleNotifications method.
func startScheduler(session *discordgo.Session) {
	startTime := time.Now().UTC()
	for {
		utils.Log.Info.WithPrefix("SCHED").Info("scheduling notifications for today's streams...")
		// check for streams tomorrow that have no time so I can be alerted to add a time
		var s db.Streams
		if tomorrowErr := s.CheckTomorrow(); tomorrowErr != nil {
			utils.Log.ErrorWarn.WithPrefix("SCHED").Error("error checking tomorrow's streams", "err", tomorrowErr)
			reports.DM(utils.Session, fmt.Sprintf("error checking tomorrow's streams:\n\terr=%s", tomorrowErr))
		}
		if len(s.Streams) > 0 {
			utils.Log.Info.WithPrefix("SCHED").Info("streams tomorrow with no time", "streams", s.Streams)
			reports.DM(utils.Session, fmt.Sprintf("streams tomorrow with no time:\n\tstreams=%v", s.Streams))
		}
		// schedule notifications for today's streams
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
		reports.DM(utils.Session, fmt.Sprintf("we survived another day\n\tuptime=%s\n\tservers=%d", time.Now().UTC().Sub(startTime).Round(time.Second), stats.GetGuildNumber(session)))
	}
}
