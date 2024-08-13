package config

// Bot is a struct that holds the bot configuration values.
type Bot struct {
	// The current version of the bot.
	Version string `toml:"version"`
	// The date the current version of the bot was released.
	ReleaseDate string `toml:"release_date"`
	// Flag to determine if the bot should restore the database from a backup.
	RestoreDatabase bool `toml:"restore_database"`
}
