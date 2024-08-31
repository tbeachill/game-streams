/*
threads.go contains functions that are run on a schedule in their own threads.
These functions perform maintenance tasks, update streams, schedule notifications,
and backup the database.
*/
package bot

import (
	"github.com/bwmarrin/discordgo"

	"gamestreams/backup"
	"gamestreams/db"
	"gamestreams/logs"
	"gamestreams/servers"
	"gamestreams/streams"
)

// streamUpdater updates the streams in the database from a web-hosted toml file.
func streamUpdater() {
	var s db.Streams
	logs.LogInfo("UPDAT", "checking for stream updates...", false)

	if updateErr := s.Update(); updateErr != nil {
		logs.LogError("UPDAT", "error updating streams",
			"err", updateErr)
	}
}

// streamNotifications schedules stream notifications for the day. The day is the
// 24-hour period between cron jobs.
func streamNotifications(session *discordgo.Session) {
	logs.LogInfo("NOTIF", "scheduling stream notifications...", false)

	if scheduleErr := streams.ScheduleNotifications(session); scheduleErr != nil {
		logs.LogError("NOTIF", "error scheduling today's streams",
			"err", scheduleErr)
	}
}

// checkTimelessStreams checks for streams that have no time set and logs them.
// a DM is also sent to the owner as a reminder to set times for the streams.
func checkTimelessStreams() {
	var s db.Streams
	if tomorrowErr := s.CheckTimeless(); tomorrowErr != nil {
		logs.LogError("TMRW ", "error checking timeless streams",
			"err", tomorrowErr)
	}
	if len(s.Streams) > 0 {
		cleanStreams := make(map[int]string)
		for _, stream := range s.Streams {
			cleanStreams[stream.ID] = stream.Name
		}
		logs.LogInfo("TMRW ", "upcoming streams with no time", true,
			"streams", cleanStreams)
	}
}

// performMaintenance performs database maintenance, clean up of logs
// blacklisted items and suggestions.
func performMaintenance(session *discordgo.Session) {
	logs.LogInfo("MNTNC", "truncating logs...", false)
	logs.TruncateLogs()
	logs.LogInfo("MNTNC", "performing server maintenance...", false)
	servers.ServerMaintenance(session)
	logs.LogInfo("MNTNC", "performing stream maintenance...", false)
	streams.StreamMaintenance()
	logs.LogInfo("MNTNC", "performing suggestion maintenance...", false)
	db.ArchiveSuggestions()
	db.RemoveOldSuggestions()
	db.PerformCommandMaintenance()
}

// backupDatabase backs up the database to a cloudflare R2 storage bucket.
func backupDatabase() {
	logs.LogInfo("BCKUP", "backing up database...", false)
	backup.BackupDB()
}
