package config

// Blacklist is a struct that holds the blacklist configuration values for the bot.
type Blacklist struct {
	// How many commands a user can run in an hour before being blacklisted.
	HourlyCommandLimit int `toml:"hourly_command_limit"`
	// How many commands a user can run in a day before being blacklisted.
	DailyCommandLimit int `toml:"daily_command_limit"`
	// How many days must pass before another message is sent to a blacklisted user.
	DaysBetweenMessages int `toml:"days_between_messages"`
}
