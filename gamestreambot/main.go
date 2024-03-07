package main

import (
	"os"

	"github.com/joho/godotenv"

	"gamestreambot/bot"
	"gamestreambot/db"
	"gamestreambot/utils"
)

// TODO: write function to DM me when there was a stream in a previous year

func main() {
	utils.SetConfig()
	utils.SetLogger()
	utils.Logger.WithPrefix(" MAIN").Info("starting GameStreamBot")

	if envErr := godotenv.Load(utils.DotEnvFile); envErr != nil {
		utils.EWLogger.WithPrefix(" MAIN").Error("error loading .env file", "err", envErr)
		os.Exit(1)
	}
	createErr := db.CreateDB()
	if createErr != nil {
		utils.EWLogger.WithPrefix(" MAIN").Error("error creating database", "err", createErr)
		os.Exit(1)
	}
	botToken := os.Getenv("DISCORD_TOKEN")
	appID := os.Getenv("APPLICATION_ID")
	bot.Run(botToken, appID)
}
