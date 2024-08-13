/*
scheduling.go contains logic to schedule functions to run at specified intervals.
*/
package bot

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"

	"gamestreams/config"
)

// ScheduleFunctions schedules the functions that need to be run on a schedule.
// It uses the cron package to schedule the functions at the intervals specified
// in the config.toml file.
func ScheduleFunctions(session *discordgo.Session) {
	c := cron.New(cron.WithLocation(time.UTC))

	if config.Values.Schedule.StreamUpdate.Enabled {
		c.AddFunc(config.Values.Schedule.StreamUpdate.Cron, func() {
			streamUpdater()
		})
	}
	if config.Values.Schedule.StreamNotifications.Enabled {
		c.AddFunc(config.Values.Schedule.StreamNotifications.Cron, func() {
			streamNotifications(session)
		})
	}
	if config.Values.Schedule.CheckTimelessStreams.Enabled {
		c.AddFunc(config.Values.Schedule.CheckTimelessStreams.Cron, func() {
			checkTimelessStreams()
		})
	}
	if config.Values.Schedule.Maintenance.Enabled {
		c.AddFunc(config.Values.Schedule.Maintenance.Cron, func() {
			performMaintenance(session)
		})
	}
	if config.Values.Schedule.Backup.Enabled {
		c.AddFunc(config.Values.Schedule.Backup.Cron, func() {
			backupDatabase()
		})
	}
	c.Start()
}
