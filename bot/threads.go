package bot

import (
	"github.com/bwmarrin/discordgo"

	"gamestreams/backup"
	"gamestreams/db"
	"gamestreams/logs"
	"gamestreams/servers"
	"gamestreams/streams"
)

// streamUpdater updates the streams in the database from a toml file
func streamUpdater() {
	var s db.Streams
	logs.LogInfo("UPDAT", "checking for stream updates...", false)

	if updateErr := s.Update(); updateErr != nil {
		logs.LogError("UPDAT", "error updating streams",
			"err", updateErr)
	}
}

// streamNotifications schedules stream notifications for the day
func streamNotifications(session *discordgo.Session) {
	logs.LogInfo("NOTIF", "scheduling stream notifications...", false)

	if scheduleErr := streams.ScheduleNotifications(session); scheduleErr != nil {
		logs.LogError("NOTIF", "error scheduling today's streams",
			"err", scheduleErr)
	}
}

// checkTomorrowsStreams checks for streams scheduled for tomorrow that do not have a time
// set. It notifies the owner of the streams that are missing a time.
func checkTomorrowsStreams() {
	var s db.Streams
	if tomorrowErr := s.CheckTomorrow(); tomorrowErr != nil {
		logs.LogError("TMRW ", "error checking tomorrow's streams",
			"err", tomorrowErr)
	}
	if len(s.Streams) > 0 {
		cleanStreams := make(map[int]string)
		for _, stream := range s.Streams {
			cleanStreams[stream.ID] = stream.Name
		}
		logs.LogInfo("TMRW ", "streams tomorrow with no time", true,
			"streams", cleanStreams)
	}
}

// performMaintenance performs database maintenance, clean up of logs
// blacklisted items and suggestions
func performMaintenance(session *discordgo.Session) {
	logs.LogInfo("MNTNC", "truncating logs...", false)
	logs.TruncateLogs()
	logs.LogInfo("MNTNC", "performing server maintenance...", false)
	servers.ServerMaintenance(session)
	logs.LogInfo("MNTNC", "performing stream maintenance...", false)
	streams.StreamMaintenance()
	logs.LogInfo("MNTNC", "performing suggestion maintenance...", false)
	db.RemoveOldSuggestions()
	logs.LogInfo("MNTNC", "deleting expired blacklisted items...", false)
	db.RemoveExpiredBlacklist()
}

// backupDatabase backs up the database to cloudflare
func backupDatabase() {
	logs.LogInfo("BCKUP", "backing up database...", false)
	backup.BackupDB()
}
