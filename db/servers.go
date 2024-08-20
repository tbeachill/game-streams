/*
servers.go contains functions that interact with the servers table of the database.
*/
package db

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"gamestreams/config"
	"gamestreams/logs"
)

// Server represents a row in the servers table of the database.
type Server struct {
	// The Discord ID of the server.
	ID string
	// The name of the server.
	Name string
	// The Discord ID of the server owner.
	OwnerID string
	// The date the bot joined the server.
	DateJoined string
	// The number of members in the server.
	MemberCount int
	// The locale of the server.
	Locale string
	// The settings for the stream announcements in the server.
	Settings Settings
}

// GetAllServerIDs returns a slice of all server IDs from the servers table
func GetAllServerIDs() ([]string, error) {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
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
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
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

// RemoveServer removes the given server ID from the servers table.
func RemoveServer(serverID string) error {
	logs.LogInfo("   DB", "removing server from servers table", false,
		"serverID", serverID)
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
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
func NewServer(serverID string, serverName string, ownerID string, joinedAt time.Time, memberCount int, locale string) error {
	logs.LogInfo("   DB", "adding new server to servers table", false,
		"serverID", serverID)

	s := Server{
		ID:          serverID,
		Name:        serverName,
		OwnerID:     ownerID,
		DateJoined:  joinedAt.UTC().Format("2006-01-02"),
		MemberCount: memberCount,
		Locale:      locale,
		Settings:    NewSettings(serverID),
	}
	if s.Set() != nil {
		return s.Set()
	}
	if s.Settings.Set() != nil {
		return s.Settings.Set()
	}
	return nil
}

// CheckServerColumns checks for servers that have missing columns in the servers table
// and returns a slice of server IDs that have missing columns.
func CheckServerColumns() ([]string, error) {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return nil, openErr
	}
	defer db.Close()

	rows, execErr := db.Query(`SELECT server_id
								FROM servers
								WHERE owner_id = ""
								OR date_joined = ""
								OR server_name = ""
								OR member_count = 0
								OR locale = ""`)

	if execErr != nil {
		return nil, execErr
	}
	defer rows.Close()

	var servers []string
	for rows.Next() {
		var serverID string
		scanErr := rows.Scan(&serverID)
		if scanErr != nil {
			return nil, scanErr
		}
		servers = append(servers, serverID)
	}

	return servers, nil
}

// Set writes the server information from the struct to the servers table in the
// database. If the server is not in the table, it will insert a new row. If the server
// is in the table, it will update the row.
func (s *Server) Set() error {
	logs.LogInfo("   DB", "setting server settings", false,
		"serverID", s.ID)
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	inServerTable, checkErr := CheckServerID(s.ID)
	if checkErr != nil {
		return checkErr
	}
	if !inServerTable {
		db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
		if openErr != nil {
			return openErr
		}
		defer db.Close()

		_, execErr := db.Exec(`INSERT INTO servers (server_id, server_name, owner_id, date_joined, member_count, locale)
								VALUES (?, ?, ?, ?, ?, ?)`,
			s.ID,
			s.Name,
			s.OwnerID,
			s.DateJoined,
			s.MemberCount,
			s.Locale)

		return execErr
	} else {
		_, execErr := db.Exec(`UPDATE servers
								SET server_name = ?,
									owner_id = ?,
									date_joined = ?,
									member_count = ?,
									locale = ?
								WHERE server_id = ?`,
			s.Name,
			s.OwnerID,
			s.DateJoined,
			s.MemberCount,
			s.Locale,
			s.ID)

		return execErr
	}
}

// Get populates the struct with information from the servers table in the database.
// It uses the server ID from the struct to query the database.
func (s *Server) Get() error {
	logs.LogInfo("   DB", "getting server settings", false,
		"serverID", s.ID)
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	row := db.QueryRow(`SELECT server_name,
							owner_id,
							date_joined,
							member_count,
							locale
						FROM servers
						WHERE server_id = ?`,
		s.ID)

	scanErr := row.Scan(&s.Name, &s.OwnerID, &s.DateJoined, &s.MemberCount, &s.Locale)
	if scanErr != nil {
		return scanErr
	}

	if getErr := s.Settings.Get(s.ID); getErr != nil {
		return getErr
	}
	return nil
}
