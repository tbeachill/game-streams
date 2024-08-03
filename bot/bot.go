package bot

import (
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/commands"
	"gamestreams/discord"
	"gamestreams/logs"
	"gamestreams/servers"
	"gamestreams/utils"
)

// Run is the main function that runs the bot. It creates a new Discord session,
// registers the commands, and registers the scheduled functions.
// The bot runs until it receives a termination signal (ctrl + c).
func Run(botToken, appID string) {
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

	servers.MonitorGuilds(session)
	utils.StartTime = time.Now().UTC()
	logs.LogInfo(" MAIN", "bot started", true)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
