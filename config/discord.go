package config

// Discord is a struct that holds the Discord configuration values for the bot.
type Discord struct {
	Token         string `toml:"token"`
	ApplicationID string `toml:"application_id"`
	OwnerID       string `toml:"owner_id"`
}
