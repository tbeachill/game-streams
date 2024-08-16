package config

// Discord is a struct that holds the credentials for the bot to interact with
// the Discord API.
type Discord struct {
	Token         string `toml:"token"`
	ApplicationID string `toml:"application_id"`
	OwnerID       string `toml:"owner_id"`
	EmbedColor    int    `toml:"embed_color"`
}
