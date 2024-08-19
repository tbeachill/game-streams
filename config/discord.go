package config

// Discord is a struct that holds the credentials for the bot to interact with
// the Discord API and some other configuration values.
type Discord struct {
	// The token for the bot to use to authenticate with the Discord API.
	Token string `toml:"token"`
	// The application ID of the bot.
	ApplicationID string `toml:"application_id"`
	// The owner of the bot's Discord user ID.
	OwnerID string `toml:"owner_id"`
	// The colour on the left side of the embeds.
	EmbedColour int `toml:"embed_colour"`
}
