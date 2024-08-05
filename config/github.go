package config

// Github is a struct that holds the Github configuration values for the bot.
type Github struct {
	StreamsTOMLURL string `toml:"streams_toml_url"`
	APIURL         string `toml:"api_url"`
}
