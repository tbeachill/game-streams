package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/utils"
)

// Server is a struct that contains information about a server. It contains the server
// ID, the name of the server, the owner ID of the server, the date the bot joined the
// server, the number of times the bot has been used in the server, and the settings for
// the server.
type Server struct {
	ID         string
	Name       string
	OwnerID    string
	DateJoined string
	UsageCount int
	Settings   Settings
}

// GetAllServerIDs returns a list of all server IDs from the servers table
func GetAllServerIDs() ([]string, error) {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return nil, openErr
	}
	defer db.Close()

	rows, queryErr := db.Query(`SELECT server_id
								FROM servers`)

	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()

	var serverIDs []string
	for rows.Next() {
		var serverID string
		scanErr := rows.Scan(&serverID)
		if scanErr != nil {
			return nil, scanErr
		}
		serverIDs = append(serverIDs, serverID)
	}
	return serverIDs, nil
}

// CheckServerID checks if the given server ID exists in the servers table. Returns
// true if the server ID exists, false if it does not.
func CheckServerID(serverID string) (bool, error) {
	println("CHECK SERVER IDS")
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return false, openErr
	}
	defer db.Close()

	row := db.QueryRow(`SELECT server_id
						FROM servers
						WHERE server_id = ?`,
		serverID)

	var checkID string
	scanErr := row.Scan(&checkID)
	if scanErr == sql.ErrNoRows {
		return false, nil
	} else if scanErr != nil {
		return false, scanErr
	}
	return true, nil
}

// GetPlatformServerIDs returns a list of server IDs that have the given platform set
// to true in the servers table.
func GetPlatformServerIDs(platform string) ([]string, error) {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return nil, openErr
	}
	defer db.Close()

	platform = strings.ToLower(platform)
	utils.Log.Info.WithPrefix(" DB").Info("getting server IDs for",
		"platform", platform)

	rows, queryErr := db.Query(`SELECT server_id
								FROM servers
								WHERE ? = ?`,
		platform,
		"1")

	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()

	var serverIDs []string
	for rows.Next() {
		var serverID string
		scanErr := rows.Scan(&serverID)
		if scanErr != nil {
			return nil, scanErr
		}
		serverIDs = append(serverIDs, fmt.Sprint(serverID))
	}
	err := rows.Err()
	if err != nil {
		utils.Log.Info.Error(err)
	}
	return serverIDs, nil
}

// RemoveServer removes the given server ID from the servers table.
func RemoveServer(serverID string) error {
	utils.LogInfo("OWNER", "removing server from servers table", false,
		"serverID", serverID)
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`DELETE FROM servers
							WHERE server_id = ?`,
		serverID)
	return execErr
}

// NewServer adds a new server to the servers table in the database.
func NewServer(serverID string, serverName string, ownerID string) error {
	utils.LogInfo("MAIN", "adding new server to servers table", false,
		"serverID", serverID)

	s := Server{
		ID:         serverID,
		Name:       serverName,
		OwnerID:    ownerID,
		DateJoined: time.Now().UTC().Format("2006-01-02"),
		UsageCount: 0,
		Settings:   NewSettings(serverID),
	}
	if s.Set() != nil {
		return s.Set()
	}
	if s.Settings.Set() != nil {
		return s.Settings.Set()
	}
	return nil
}

// IncrementUsageCount increments the usage count for the given server ID in the servers
// table.
func IncrementUsageCount(serverID string) error {
	utils.LogInfo("OWNER", "incrementing server usage count", false,
		"serverID", serverID)
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`UPDATE servers
							SET usage_count = usage_count + 1
							WHERE server_id = ?`,
		serverID)
	return execErr
}

// check for servers that are not in the servers table
func checkServers(servers []string) error {
	for _, serverID := range servers {
		if exists, checkErr := CheckServerID(serverID); checkErr != nil {
			return checkErr
		} else if !exists {
			if newErr := NewServer(serverID, "", ""); newErr != nil {
				return newErr
			}
		}
	}
	return nil
}

// check for servers that have missing columns in the servers table
func checkServerColumns() error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Query(`SELECT server_id,
							FROM servers
							WHERE owner_id IS NULL
							OR date_joined IS NULL`)
	if execErr != nil {
		return execErr
	}
	return nil
}

// Set sets the settings for the given server ID in the servers table.
func (s *Server) Set() error {
	utils.LogInfo("OWNER", "setting server settings", false,
		"serverID", s.ID)
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`INSERT INTO servers
							(server_id,
							server_name,
							owner_id,
							date_joined,
							usage_count)
						VALUES (?,?, ?, ?, ?)`,
		s.ID,
		s.Name,
		s.OwnerID,
		s.DateJoined,
		s.UsageCount)
	if execErr != nil {
		return execErr
	}
	return nil
}
