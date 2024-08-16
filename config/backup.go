package config

// Backup is a struct that holds the backup configuration values for the bot.
type Backup struct {
	// The number of days to keep backups for.
	DaysToKeep int `toml:"days_to_keep"`
}
