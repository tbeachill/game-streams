package config

// Schedules is a struct that holds the schedules for the bot.
type Schedules struct {
	// The schedule for performing a backup of the database.
	Backup Schedule `toml:"backup"`
	// The schedule for performaing a range of maintenance tasks.
	Maintenance Schedule `toml:"maintenance"`
	// The schedule for updating the stream data from the TOML file.
	StreamUpdate Schedule `toml:"stream_update"`
	// The schedule for queueing stream notifications
	StreamNotifications Schedule `toml:"stream_notifications"`
	// The schedule for checking streams with no time set
	CheckTimelessStreams Schedule `toml:"timeless_streams"`
	// The number of minutes before a stream starts to send a notification.
	NotificationTMinus int `toml:"notification_t_minus"`
}

// Schedule is a struct that holds the configuration values for each schedule.
type Schedule struct {
	Enabled bool   `toml:"enabled"`
	Cron    string `toml:"cron"`
}
