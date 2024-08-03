package bot

import (
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/backup"
	"gamestreams/db"
	"gamestreams/logs"
	"gamestreams/servers"
	"gamestreams/streams"

)

// startUpdater creates a new Streams struct and updates the streams table of the database
// at scheduled intervals.
func streamUpdater() {
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

// startScheduler runs at startup, then at scheduled intervals. It schedules notifications
// for today's streams, then sleeps until the next day to schedule notifications for the
// next day.
func streamNotifications(session *discordgo.Session) {
	// timeToRun is the hour of the day in UTC to run the scheduler
	timeToRun := 6

	for {
		logs.LogInfo("SCHED", "scheduling stream notifications...", false)
		// check for streams tomorrow that have no time so I can add a time

		// schedule notifications for today's streams
		if scheduleErr := streams.ScheduleNotifications(session); scheduleErr != nil {
			logs.LogError("SCHED", "error scheduling today's streams",
				"err", scheduleErr)
		}

		hour, min, _ := time.Now().UTC().Clock()
		hoursRemaining := (timeToRun + 24) - hour
		minsRemaining := 60 - min
		logs.LogInfo("SCHED", "sleeping until next day", false,
			"hours", hoursRemaining,
			"minutes", minsRemaining)
		time.Sleep(time.Duration(hoursRemaining*60+minsRemaining) * time.Minute)
	}
}

// checkTomorrowsStreams runs at startup, then at scheduled intervals. It checks if there
// are any streams tomorrow with no time set and alerts the owner to add a time.
func checkTomorrowsStreams() {
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
}

// performMaintenance runs at startup, then at scheduled intervals. It truncates the logs,
// backs up the database, performs database maintenance, and deletes expired blacklisted
// items by running the corresponding functions.
func performMaintenance(session *discordgo.Session) {
	logs.LogInfo("MAINT", "truncating logs...", false)
	logs.TruncateLogs()
	logs.LogInfo("MAINT", "backing up database...", false)
	backup.BackupDB()
	logs.LogInfo("MAINT", "performing server maintenance...", false)
	servers.ServerMaintenance(session)
	logs.LogInfo("MAINT", "performing stream maintenance...", false)
	streams.StreamMaintenance()
	logs.LogInfo("MAINT", "deleting expired blacklisted items...", false)
	db.RemoveExpiredBlacklist()
}
