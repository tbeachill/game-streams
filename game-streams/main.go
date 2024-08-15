/*
main.go is the entry point of the program.
*/
package main

import (
	"os"

	"gamestreams/bot"
	"gamestreams/config"
	"gamestreams/db"
	"gamestreams/logs"
)

// main loads the configuration values from config.toml, initialises the logs, creates
// the database, and starts the bot.
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
