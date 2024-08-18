package config

// Suggestions is a struct that holds the configuration values for the suggestions.
type Suggestions struct {
	// The limit of the number of suggestions returned from the /suggestions command.
	DaysToKeep int `toml:"days_to_keep"`
	// The number of suggestions to allow per user per day.
	DailyLimit int `toml:"daily_limit"`
}
