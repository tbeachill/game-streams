package config

// Commands is a struct that holds information about the commands table of the database.
type Commands struct {
	// The number of months to keep command data in the database.
	MonthsToKeep int `toml:"months_to_keep"`
}
