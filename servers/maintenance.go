package servers

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/db"
	"gamestreams/discord"
	"gamestreams/logs"
)

// ServerMaintenance performs maintenance on the servers table.
// It gets the list of servers the bot is in from the Discord API. It then
// leaves blacklisted servers, adds servers that are in the Discord list but
// not in the servers table, removes servers that are in the table but not in
// the Discord list, and adds missing columns to the servers table.
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

			newErr := db.NewServer(server.ID, server.Name, server.OwnerID, server.JoinedAt, server.MemberCount, server.PreferredLocale)
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

// updateMemberCount updates the member count of a server from its ID in the servers
// table.
func updateMemberCount(serverID string) (int, error) {
	server, err := discord.Session.Guild(serverID)
	if err != nil {
		logs.LogError("SERVR", "error getting server",
			"err", err)
		return 0, err
	}
	return server.MemberCount, nil
}

// getDateJoined returns the date a server was joined by the bot from its ID.
func getDateJoined(serverID string) (string, error) {
	server, err := discord.Session.Guild(serverID)
	if err != nil {
		return "", err
	}
	dateJoined := strings.Split(server.JoinedAt.Format(time.RFC3339), "T")[0]
	return dateJoined, nil
}

// getServerLocale returns the locale of a server from its ID.
func getServerLocale(serverID string) (string, error) {
	server, err := discord.Session.Guild(serverID)
	if err != nil {
		return "", err
	}
	return server.PreferredLocale, nil
}
