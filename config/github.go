package config

// Github is a struct that holds the Github configuration values for the bot.
type Github struct {
	StreamDataURL string `toml:"stream_data_url"`
	APIURL        string `toml:"api_url"`
}
