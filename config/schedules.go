package config

// Schedule is a struct that holds the schedule configuration values for the bot.
type Schedules struct {
	Backup                Schedule `toml:"backup"`
	Maintenance           Schedule `toml:"maintenance"`
	StreamUpdate          Schedule `toml:"stream_update"`
	StreamNotifications   Schedule `toml:"stream_notifications"`
	CheckTomorrowsStreams Schedule `toml:"tomorrows_streams"`
	NotificationTMinus    int      `toml:"notification_t_minus"`
}

// Schedule is a struct that holds the schedule configuration values for the bot.
type Schedule struct {
	Enabled bool   `toml:"enabled"`
	Cron    string `toml:"cron"`
}
