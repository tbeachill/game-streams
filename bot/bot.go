package bot

import (
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/backup"
	"gamestreams/commands"
	"gamestreams/db"
	"gamestreams/discord"
	"gamestreams/logs"
	"gamestreams/servers"
	"gamestreams/streams"
	"gamestreams/utils"
)

// Run is the main function that runs the bot. It creates a new Discord session,
// registers the commands, and starts the updater and scheduler goroutines.
// The bot runs until it receives a termination signal (ctrl + c).
func Run(botToken, appID string) {
	session, sessionErr := discordgo.New("Bot " + botToken)
	if sessionErr != nil {
		logs.LogError(" MAIN", "error creating Discord session",
			"err", sessionErr)
		return
	}
	if openErr := session.Open(); openErr != nil {
		logs.LogError(" MAIN", "error connecting to Discord",
			"err", openErr)
		return
	}
	defer session.Close()
	// register the session with the logs and discord packages
	logs.RegisterSession(session)
	discord.RegisterSession(session)
	//commands.RemoveAllCommands(appID, session)
	commands.RegisterCommands(appID, session)
	commands.RegisterHandler(session, &discordgo.InteractionCreate{})
	commands.RegisterOwnerCommands(session)
	go startUpdater()
	go startScheduler(session)
	servers.MonitorGuilds(session)
	utils.StartTime = time.Now().UTC()
	logs.LogInfo(" MAIN", "bot started", true)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

// startUpdater creates a new Streams struct and updates the streams every 6 hours by
// running the Update method.
func startUpdater() {
	var s db.Streams
	for {
		logs.LogInfo("UPDAT", "checking for stream updates...", false)
		if updateErr := s.Update(); updateErr != nil {
			logs.LogError("UPDAT", "error updating streams",
				"err", updateErr)
		}
		hoursRemaining := 6 - ((time.Now().UTC().Hour()) % 6)
		logs.LogInfo("UPDAT", "sleeping until next update", false,
			"hours", hoursRemaining)

		time.Sleep(time.Duration(hoursRemaining) * time.Hour)
	}
}

// startScheduler runs at startup, then at midnight UTC every day. It checks if there
// are any streams tomorrow with no time set and alerts me to add a time. It creates a
// streams struct with today's streams and schedules notifications for each stream by
// running the ScheduleNotifications method.
func startScheduler(session *discordgo.Session) {
	// timeToRun is the hour of the day in UTC to run the scheduler
	timeToRun := 6

	for {
		logs.LogInfo("SCHED", "running scheduler...", false)
		// check for streams tomorrow that have no time so I can add a time
		var s db.Streams
		if tomorrowErr := s.CheckTomorrow(); tomorrowErr != nil {
			logs.LogError("SCHED", "error checking tomorrow's streams",
				"err", tomorrowErr)
		}
		if len(s.Streams) > 0 {
			cleanStreams := make(map[int]string)
			for _, stream := range s.Streams {
				cleanStreams[stream.ID] = stream.Name
			}
			logs.LogInfo("SCHED", "streams tomorrow with no time", true,
				"streams", cleanStreams)
		}
		// schedule notifications for today's streams
		if scheduleErr := streams.ScheduleNotifications(session); scheduleErr != nil {
			logs.LogError("SCHED", "error scheduling today's streams",
				"err", scheduleErr)
		}
		logs.LogInfo("SCHED", "truncating logs...", false)
		utils.TruncateLogs()
		logs.LogInfo("SCHED", "backing up database...", false)
		backup.BackupDB()
		logs.LogInfo("SCHED", "performing server maintenance...", false)
		servers.ServerMaintenance(session)
		logs.LogInfo("SCHED", "performing stream maintenance...", false)
		streams.StreamMaintenance()
		logs.LogInfo("SCHED", "deleting expired blacklisted items...", false)
		db.RemoveExpiredBlacklist()

		hour, min, _ := time.Now().UTC().Clock()
		hoursRemaining := (timeToRun + 24) - hour
		minsRemaining := 60 - min
		logs.LogInfo("SCHED", "sleeping until next day", false,
			"hours", hoursRemaining,
			"minutes", minsRemaining)
		time.Sleep(time.Duration(hoursRemaining*60+minsRemaining) * time.Minute)
	}
}
