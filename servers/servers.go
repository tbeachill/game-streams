package servers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"gamestreambot/db"
	"gamestreambot/utils"
)

// GetGuildNumber returns the number of servers the bot is in.
func GetGuildNumber(session *discordgo.Session) int {
	num := len(session.State.Guilds)
	return num
}

// logGuildNumber reports the number of servers the bot is in to the console and the
// bot owner via DM.
func logGuildNumber(session *discordgo.Session) {
	guildNum := GetGuildNumber(session)
	utils.LogInfo("SERVR", "connected", true,
		"server_count", guildNum)
}

// MonitorGuilds monitors the servers the bot is in. It sets up handlers for when the
// bot joins or leaves a server. When the bot joins a server, it checks if the server is
// already in the servers table of the database. If not, it adds the server to the
// servers table with default options. When the bot is removed from a server, it
// removes the server from the servers table.
func MonitorGuilds(session *discordgo.Session) {
	logGuildNumber(session)
	utils.LogInfo("SERVR", "adding server join handler", false)

	// join handler
	session.AddHandler(func(s *discordgo.Session, e *discordgo.GuildCreate) {
		utils.LogInfo("SERVR", "joined server", true,
			"server", e.Guild.Name,
			"owner", e.Guild.OwnerID)
		logGuildNumber(s)
		// check if server is blacklisted
		if leaveErr := LeaveIfBlacklisted(s, e.Guild.ID, e); leaveErr != nil {
			utils.LogError("SERVR", "error leaving blacklisted server",
				"server", e.Guild.Name,
				"err", leaveErr)
			return
		}
		// check if server ID is in the servers table
		present, checkErr := db.CheckServerID(e.Guild.ID)
		if checkErr != nil {
			utils.LogError("SERVR", "error checking server ID",
				"err", checkErr)
			return
		}
		if !present {
			utils.LogInfo("SERVR", "adding server to database", false,
				"server", e.Guild.Name)

			utils.IntroDM(e.OwnerID)

			newErr := db.NewServer(e.Guild.ID, e.Guild.Name, e.Guild.OwnerID)
			if newErr != nil {
				utils.LogError("SERVR", "error adding server to database",
					"server", e.Guild.Name,
					"err", newErr)
			}
		}
	})
	utils.LogInfo("SERVR", "adding server leave handler", false)

	// leave handler
	session.AddHandler(func(s *discordgo.Session, e *discordgo.GuildDelete) {
		utils.LogInfo("SERVR", "left server", true,
			"server", e.Guild.Name,
			"owner", e.Guild.OwnerID)

		logGuildNumber(s)
		if removeErr := db.RemoveServer(e.Guild.ID); removeErr != nil {
			utils.LogError("SERVR", "error removing server",
				"server", e.Guild.Name,
				"err", removeErr)
		}
	})
}

// GetAllServerIDsFromDiscord returns a slice of all server IDs the bot is in as
// returned by Discord.
func GetAllServerIDsFromDiscord(session *discordgo.Session) []string {
	var serverIDs []string
	for _, guild := range session.State.Guilds {
		serverIDs = append(serverIDs, guild.ID)
	}
	return serverIDs
}

// RemoveOldServerIDs removes server IDs from the servers table that are not in the
// Discord returned list of server IDs. This is for data cleanup in case the bot is
// removed from a server and the server ID is not removed from the database.
func RemoveOldServerIDs(session *discordgo.Session) error {
	discordServerIDs := GetAllServerIDsFromDiscord(session)
	dbServerIDs, getErr := db.GetAllServerIDs()

	if getErr != nil {
		return getErr
	}
	for _, dbID := range dbServerIDs {
		found := false
		for _, discordID := range discordServerIDs {
			if dbID == discordID {
				found = true
				break
			}
		}
		if !found {
			utils.LogInfo("SERVR", "removing old server ID", false,
				"server", dbID)

			if removeErr := db.RemoveServer(dbID); removeErr != nil {
				return removeErr
			}
		}
	}
	return nil
}

// leaveServer leaves the server with the given server ID. It sends a DM to the server
// owner with the reason the bot left the server.
func leaveServer(session *discordgo.Session, serverID string, reason string, e *discordgo.GuildCreate) error {
	utils.LogInfo("SERVR", "leaving server", true,
		"server", serverID,
		"reason", reason)

	utils.DM(e.OwnerID, fmt.Sprintf("I am leaving your server because %s", reason))
	if removeErr := session.GuildLeave(serverID); removeErr != nil {
		return removeErr
	}

	return nil
}

// leaveIfBlacklisted checks if the server with the given server ID is blacklisted. If
// it is, the bot leaves the server.
func LeaveIfBlacklisted(session *discordgo.Session, serverID string, e *discordgo.GuildCreate) error {
	blacklisted, reason := db.IsBlacklisted(serverID, "server")
	if blacklisted {
		return leaveServer(session, serverID, fmt.Sprintf("blacklisted: %s", reason), e)
	}
	return nil
}

// GetServerName returns the name of a server from a server ID.
func GetServerName(serverID string) string {
	server, err := utils.Session.Guild(serverID)
	if err != nil {
		utils.LogError(" MAIN", "error getting server name", "err", err)
		return ""
	}
	return server.Name
}

// GetServerOwner returns the owner of a server from a server ID.
func GetServerOwner(serverID string) string {
	server, err := utils.Session.Guild(serverID)
	if err != nil {
		utils.LogError(" MAIN", "error getting server owner", "err", err)
		return ""
	}
	return server.OwnerID
}

// ServerMaintenance checks the servers table for servers that are no longer in the
// discord returned list of servers and removes them from the servers table.
// it then checks for connected servers that are not in the servers table and adds
// them to the servers table. Then it checks for blacklisted servers and removes them
// from the servers table. Then it checks for servers in the servers table with missing
// columns and adds the missing columns.
func ServerMaintenance(session *discordgo.Session) error {
	// remove servers that are not in the discord list
	// add servers that are in the discord list but not in the servers table
	// remove blacklisted servers
	// add missing columns
}
