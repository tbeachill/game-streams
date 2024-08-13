/*
bot.go contains the main function that runs the bot.
It is the first file called by main.go when the bot is started.
*/
package bot

import (
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/backup"
	"gamestreams/commands"
	"gamestreams/config"
	"gamestreams/discord"
	"gamestreams/logs"
	"gamestreams/servers"
	"gamestreams/utils"
)

// Run is the main function that runs the bot. It creates a new Discord session,
// registers the commands, and registers the scheduled functions.
// If the restore flag is set, it restores the database from the most recent backup
// then exits. The bot runs until it receives a termination signal (ctrl + c).
func Run(botToken, appID string) {
	if config.Values.Bot.RestoreDatabase {
		backup.BackupDB()
		logs.LogInfo(" MAIN", "RESTORE FLAG SET: RESTORING DATABASE", false)
		backup.RestoreDB()
		os.Exit(0)
	}

	session, sessionErr := discordgo.New("Bot " + botToken)
	if sessionErr != nil {
		logs.LogError(" MAIN", "error creating Discord session",
			"err", sessionErr)
		return
	}
	if openErr := session.Open(); openErr != nil {
		logs.LogError(" MAIN", "error connecting to Discord",
			"err", openErr)
		return
	}
	defer session.Close()

	ScheduleFunctions(session)

	logs.RegisterSession(session)
	discord.RegisterSession(session)
	//commands.RemoveAllCommands(appID, session)
	commands.RegisterCommands(appID, session)
	commands.RegisterHandler(session, &discordgo.InteractionCreate{})
	commands.RegisterOwnerCommands(session)

	// Run some of the scheduled functions immediately
	streamUpdater()
	performMaintenance(session)
	streamNotifications(session)
	checkTimelessStreams()

	servers.MonitorGuilds(session)
	utils.StartTime = time.Now().UTC()
	logs.LogInfo(" MAIN", "bot started", true)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
