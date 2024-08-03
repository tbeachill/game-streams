package main

import (
	"os"

	"gamestreams/bot"
	"gamestreams/config"
	"gamestreams/db"
	"gamestreams/logs"
)

// main is the entry point for the bot. It sets the file paths, initializes the logger,
// and loads the .env file.
// It then creates/loads the database and starts the bot.
// The environment variables are:
//
//	DISCORD_TOKEN - the token for the Discord bot.
//	APPLICATION_ID - the application ID for the Discord bot.
//	OWNER_ID - the Discord user ID of the bot owner.
//	STREAM_URL - the Github URL for the streams.toml file.
//	API_URL - the Github API URL for the repository that contains the streams.toml file.
func main() {
	config.Values.Load()
	logs.Log.Init()
	logs.Log.Info.WithPrefix(" MAIN").Info("starting bot")

	createErr := db.CreateDB()
	if createErr != nil {
		logs.Log.ErrorWarn.WithPrefix(" MAIN").Error("error creating database",
			"err", createErr)
		os.Exit(1)
	}
	bot.Run(config.Values.Discord.Token, config.Values.Discord.ApplicationID)
}
