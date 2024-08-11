package config

// Blacklist is a struct that holds the blacklist configuration values for the bot.
type Blacklist struct {
	HourlyCommandLimit int `toml:"hourly_command_limit"`
	DailyCommandLimit  int `toml:"daily_command_limit"`
}
