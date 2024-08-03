package servers

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/db"
	"gamestreams/discord"
	"gamestreams/logs"
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
	logs.LogInfo("SERVR", "connected", true,
		"server_count", guildNum)
}

// MonitorGuilds monitors the servers the bot is in. It sets up handlers for when the
// bot joins or leaves a server. When the bot joins a server, it checks if the server is
// already in the servers table of the database. If not, it adds the server to the
// servers table with default options. When the bot is removed from a server, it
// removes the server from the servers table.
func MonitorGuilds(session *discordgo.Session) {
	logGuildNumber(session)
	logs.LogInfo("SERVR", "adding server join handler", false)

	// join handler
	session.AddHandler(func(s *discordgo.Session, e *discordgo.GuildCreate) {
		logs.LogInfo("SERVR", "joined server", true,
			"server", e.Guild.Name,
			"server_id", e.Guild.ID,
			"owner", e.Guild.OwnerID)
		logGuildNumber(s)
		// check if server is blacklisted
		if leaveErr := LeaveIfBlacklisted(s, e.Guild.ID, e); leaveErr != nil {
			logs.LogError("SERVR", "error leaving blacklisted server",
				"server", e.Guild.Name,
				"server_id", e.Guild.ID,
				"err", leaveErr)
			return
		}
		// check if server ID is in the servers table
		present, checkErr := db.CheckServerID(e.Guild.ID)
		if checkErr != nil {
			logs.LogError("SERVR", "error checking server ID",
				"err", checkErr)
			return
		}
		if !present {
			logs.LogInfo("SERVR", "adding server to database", false,
				"server", e.Guild.Name)

			discord.IntroDM(e.OwnerID)

			newErr := db.NewServer(e.Guild.ID, e.Guild.Name, e.Guild.OwnerID, e.Guild.MemberCount, e.Guild.PreferredLocale)
			if newErr != nil {
				logs.LogError("SERVR", "error adding server to database",
					"server", e.Guild.Name,
					"err", newErr)
			}
		}
	})
	logs.LogInfo("SERVR", "adding server leave handler", false)

	// leave handler
	session.AddHandler(func(s *discordgo.Session, e *discordgo.GuildDelete) {
		logs.LogInfo("SERVR", "left server", true,
			"server", e.Guild.Name,
			"server_id", e.Guild.ID,
			"owner", e.Guild.OwnerID)

		logGuildNumber(s)
		if removeErr := db.RemoveServer(e.Guild.ID); removeErr != nil {
			logs.LogError("SERVR", "error removing server",
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
			logs.LogInfo("SERVR", "removing old server ID", false,
				"server", dbID)
			if removeErr := db.RemoveServer(dbID); removeErr != nil {
				return removeErr
			}
			logs.LogInfo("SERVR", "removing old server settings", false,
				"server", dbID)
			if removeErr := db.RemoveServerSettings(dbID); removeErr != nil {
				return removeErr
			}
			logs.LogInfo("SERVR", "removing command data", false,
				"server", dbID)
			if removeErr := db.RemoveCommandData(dbID); removeErr != nil {
				return removeErr
			}
		}
	}
	return nil
}

// leaveServer leaves the server with the given server ID. It sends a DM to the server
// owner with the reason the bot left the server.
func leaveServer(session *discordgo.Session, serverID string, reason string, e *discordgo.GuildCreate) error {
	logs.LogInfo("SERVR", "leaving server", true,
		"server", serverID,
		"reason", reason)

	discord.DM(e.OwnerID, fmt.Sprintf("I am leaving your server because %s", reason))
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
	server, err := discord.Session.Guild(serverID)
	if err != nil {
		logs.LogError(" MAIN", "error getting server name", "err", err)
		return ""
	}
	return server.Name
}

// GetServerOwner returns the owner of a server from a server ID.
func GetServerOwner(serverID string) string {
	server, err := discord.Session.Guild(serverID)
	if err != nil {
		logs.LogError(" MAIN", "error getting server owner", "err", err)
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
func ServerMaintenance(session *discordgo.Session) {
	servers := session.State.Guilds
	// add servers that are in the discord list but not in the servers table
	// remove blacklisted servers
	for _, server := range servers {
		// check if server is blacklisted
		if leaveErr := LeaveIfBlacklisted(session, server.ID, nil); leaveErr != nil {
			logs.LogError("SERVR", "error leaving blacklisted server",
				"server", server.Name,
				"err", leaveErr)
			return
		}
		// check if server ID is in the servers table
		present, checkErr := db.CheckServerID(server.ID)
		if checkErr != nil {
			logs.LogError("SERVR", "error checking server ID",
				"err", checkErr)
			return
		}
		if !present {
			logs.LogInfo("SERVR", "adding server to database", false,
				"server", server.Name)

			newErr := db.NewServer(server.ID, server.Name, server.OwnerID, server.MemberCount, server.PreferredLocale)
			if newErr != nil {
				logs.LogError("SERVR", "error adding server to database",
					"server", server.Name,
					"err", newErr)
				return
			}
		}
	}

	// remove servers that are in the table but not in the discord list
	if removeErr := RemoveOldServerIDs(session); removeErr != nil {
		logs.LogError("SERVR", "error removing old server IDs",
			"err", removeErr)
		return
	}

	// check for servers that have missing columns in the servers table
	serverIDs, checkErr := db.CheckServerColumns()
	if checkErr != nil {
		logs.LogError("SERVR", "error checking server columns",
			"err", checkErr)
		return
	}

	// add missing columns
	if len(serverIDs) > 0 {
		logs.LogInfo("SERVR", "adding missing columns and updating member counts for servers", false,
			"servers", serverIDs)

		for _, serverID := range serverIDs {
			s := db.Server{
				ID: serverID,
			}
			s.Get()
			if s.Name == "" {
				s.Name = GetServerName(serverID)
			}
			if s.OwnerID == "" {
				s.OwnerID = GetServerOwner(serverID)
			}
			if s.DateJoined == "" {
				dateJoined, dateErr := getDateJoined(serverID)
				if dateErr != nil {
					logs.LogError("SERVR", "error getting date joined",
						"err", dateErr)
				} else {
					s.DateJoined = dateJoined
				}
			}
			if !db.CheckSettings(serverID) {
				s.Settings = db.NewSettings(serverID)
				if setErr := s.Settings.Set(); setErr != nil {
					logs.LogError("SERVR", "error setting server settings",
						"err", setErr)
				}
			}
			if s.Locale == "" {
				locale, localeErr := getServerLocale(serverID)
				if localeErr != nil {
					logs.LogError("SERVR", "error getting server locale",
						"err", localeErr)
				} else {
					s.Locale = locale
				}
			}
			memberCount, countErr := updateMemberCount(serverID)
			if countErr != nil {
				logs.LogError("SERVR", "error getting member count",
					"err", countErr)
			} else {
				s.MemberCount = memberCount
			}
			if setErr := s.Set(); setErr != nil {
				logs.LogError("SERVR", "error setting server columns",
					"err", setErr)
			}
		}
	}
}

// updateMemberCount updates the member count of a server in the servers table.
func updateMemberCount(serverID string) (int, error) {
	server, err := discord.Session.Guild(serverID)
	if err != nil {
		logs.LogError("SERVR", "error getting server",
			"err", err)
		return 0, err
	}
	return server.MemberCount, nil
}

// getDateJoined returns the date a server was joined by the bot.
func getDateJoined(serverID string) (string, error) {
	server, err := discord.Session.Guild(serverID)
	if err != nil {
		return "", err
	}
	dateJoined := strings.Split(server.JoinedAt.Format(time.RFC3339), "T")[0]
	return dateJoined, nil
}

// getServerLocale returns the locale of a server.
func getServerLocale(serverID string) (string, error) {
	server, err := discord.Session.Guild(serverID)
	if err != nil {
		return "", err
	}
	return server.PreferredLocale, nil
}
