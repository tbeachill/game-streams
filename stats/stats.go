package stats

import (
	"github.com/bwmarrin/discordgo"

	"gamestreambot/db"
	"gamestreambot/utils"
)

// returns the number of servers the bot is currently in by checking the cache
func getGuildNumber(session *discordgo.Session) int {
	num := len(session.State.Guilds)
	return num
}

// logs the number of servers the bot is in
func logGuildNumber(session *discordgo.Session) {
	guildNum := getGuildNumber(session)
	utils.Log.Info.WithPrefix("STATS").Infof("connected to %d server%s", guildNum, utils.Pluralise(guildNum))
}

// logs when the bot joins or leaves a server
func MonitorGuilds(session *discordgo.Session) {
	logGuildNumber(session)
	utils.Log.Info.WithPrefix("STATS").Info("adding server join handler")
	session.AddHandler(func(s *discordgo.Session, e *discordgo.GuildCreate) {
		utils.Log.Info.WithPrefix("STATS").Info("joined server", "server", e.Guild.Name)
		logGuildNumber(s)
		db.SetDefaultOptions(e.Guild.ID)
	})
	utils.Log.Info.WithPrefix("STATS").Info("adding server leave handler")
	session.AddHandler(func(s *discordgo.Session, e *discordgo.GuildDelete) {
		utils.Log.Info.WithPrefix("STATS").Info("left server", "server", e.Guild.Name)
		logGuildNumber(s)
		db.RemoveOptions(e.Guild.ID)
	})
}
