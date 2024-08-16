package config

// Streams is a struct that holds the configuration values for the streams.
type Streams struct {
	// The limit of the number of streams returned from the /streams command.
	Limit int `toml:"limit"`
	// The number of months to keep the streams in the database before deleting them.
	MonthsToKeep int `toml:"months_to_keep"`
}
