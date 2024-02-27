package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"gamestreambot/bot"
	"gamestreambot/db"
	"gamestreambot/utils"
)

// TODO: write function to DM me when there was a stream in a previous year

func main() {
	utils.SetConfig()

	if envErr := godotenv.Load(utils.DotEnvFile); envErr != nil {
		log.Printf("error loading .env file: %e\n", envErr)
		os.Exit(1)
	}
	createErr := db.CreateDB()
	if createErr != nil {
		log.Printf("error creating database: %e\n", createErr)
		os.Exit(1)
	}
	botToken := os.Getenv("DISCORD_TOKEN")
	appID := os.Getenv("APPLICATION_ID")
	bot.Run(botToken, appID)
}
