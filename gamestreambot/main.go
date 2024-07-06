package main

import (
	"os"

	"github.com/joho/godotenv"

	"gamestreambot/bot"
	"gamestreambot/db"
	"gamestreambot/utils"
)

// main is the entry point for the bot. It sets the file paths, initializes the logger, and loads the .env file.
// It then creates/loads the database and starts the bot.
// The environment variables are:
//
//	DISCORD_TOKEN - the token for the Discord bot.
//	APPLICATION_ID - the application ID for the Discord bot.
//	OWNER_ID - the Discord user ID of the bot owner.
//	STREAM_URL - the Github URL for the streams.toml file.
//	API_URL - the Github API URL for the repository that contains the streams.toml file.
func main() {
	utils.Files.SetPaths()
	utils.Log.Init()
	utils.Log.Info.WithPrefix(" MAIN").Info("starting bot")

	if envErr := godotenv.Load(utils.Files.DotEnv); envErr != nil {
		utils.Log.ErrorWarn.WithPrefix(" MAIN").Error("error loading .env file", "err", envErr)
		os.Exit(1)
	}
	createErr := db.CreateDB()
	if createErr != nil {
		utils.Log.ErrorWarn.WithPrefix(" MAIN").Error("error creating database", "err", createErr)
		os.Exit(1)
	}
	botToken := os.Getenv("DISCORD_TOKEN")
	appID := os.Getenv("APPLICATION_ID")
	bot.Run(botToken, appID)
}
