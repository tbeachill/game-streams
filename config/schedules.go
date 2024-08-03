package config

// Schedule is a struct that holds the schedule configuration values for the bot.
type Schedules struct {
	Maintenance  Schedule `toml:"maintenance"`
	StreamUpdate Schedule `toml:"stream_update"`
}

// Schedule is a struct that holds the schedule configuration values for the bot.
type Schedule struct {
	Enabled bool   `toml:"enabled"`
	Cron    string `toml:"cron"`
}
