package config

// Github is a struct that holds the Github configuration values for the bot.
type Github struct {
	// The URL to the streams TOML file allowing bulk import of streams.
	StreamsTOMLURL string `toml:"streams_toml_url"`
	// The URL to the Github API of the TOML file allowing commit information
	// to be retrieved.
	APIURL string `toml:"api_url"`
}
