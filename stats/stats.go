package stats

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"gamestreambot/db"
	"gamestreambot/reports"
	"gamestreambot/utils"
)

// GetGuildNumber returns the number of servers the bot is in.
func GetGuildNumber(session *discordgo.Session) int {
	num := len(session.State.Guilds)
	return num
}

// logGuildNumber reports the number of servers the bot is in to the console and the bot owner via DM.
func logGuildNumber(session *discordgo.Session) {
	guildNum := GetGuildNumber(session)
	utils.Log.Info.WithPrefix("STATS").Infof("connected to %d server%s", guildNum, utils.Pluralise(guildNum))
	reports.DM(session, fmt.Sprintf("connected to %d server%s", guildNum, utils.Pluralise(guildNum)))
}

// MonitorGuilds monitors the servers the bot is in. It sets up handlers for when the bot joins or leaves a server.
// It logs to console and DMs the bot owner when the bot joins or leaves a server.
// When the bot joins a server, it checks if the server is already in the servers table of the database. If not, it
// adds the server to the servers table with default options. When the bot is removed from a server, it removes the
// server from the servers table.
func MonitorGuilds(session *discordgo.Session) {
	logGuildNumber(session)
	utils.Log.Info.WithPrefix("STATS").Info("adding server join handler")

	// join handler
	session.AddHandler(func(s *discordgo.Session, e *discordgo.GuildCreate) {
		utils.Log.Info.WithPrefix("STATS").Info("joined server", "server", e.Guild.Name)
		reports.DM(s, fmt.Sprintf("joined server:\n\tserver=%s", e.Guild.Name))
		logGuildNumber(s)

		present, checkErr := db.CheckServerID(e.Guild.ID)
		if checkErr != nil {
			utils.Log.ErrorWarn.WithPrefix("STATS").Error("error checking server ID", "err", checkErr)
			reports.DM(s, fmt.Sprintf("error checking server ID:\n\terr=%s", checkErr))
			return
		}
		if !present {
			utils.Log.Info.WithPrefix("STATS").Info("adding server to database", "server", e.Guild.Name)
			utils.IntroDM(e.OwnerID)
			o := db.NewOptions(e.Guild.ID)
			if setErr := o.Set(); setErr != nil {
				utils.Log.ErrorWarn.WithPrefix("STATS").Error("error setting server options", "server", e.Guild.Name, "err", setErr)
				reports.DM(s, fmt.Sprintf("error setting server options:\n\tserver=%s\n\terr=%s", e.Guild.Name, setErr))
			}
		}
	})
	utils.Log.Info.WithPrefix("STATS").Info("adding server leave handler")

	// leave handler
	session.AddHandler(func(s *discordgo.Session, e *discordgo.GuildDelete) {
		utils.Log.Info.WithPrefix("STATS").Info("left server", "server", e.Guild.Name)
		reports.DM(s, fmt.Sprintf("left server:\n\tserver=%s", e.Guild.Name))
		logGuildNumber(s)
		if removeErr := db.RemoveOptions(e.Guild.ID); removeErr != nil {
			utils.Log.ErrorWarn.WithPrefix("STATS").Error("error removing server options", "server", e.Guild.Name, "err", removeErr)
			reports.DM(s, fmt.Sprintf("error removing server options:\n\tserver=%s\n\terr=%s", e.Guild.Name, removeErr))
		}
	})
}
