package config

// Suggestions is a struct that holds the configuration values for the suggestions.
type Suggestions struct {
	// The limit of the number of suggestions returned from the /suggestions command.
	DaysToKeep int `toml:"days_to_keep"`
}
