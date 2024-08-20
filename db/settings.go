/*
settings.go contains the Settings struct and functions that interact with the
server_settings table of the database.
*/
package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"gamestreams/config"
	"gamestreams/logs"
)

// Settings is a struct that contains the settings for a server. These settings are used
// to determine which streams to announce in a server, where to announce them, and which
// roles to ping when announcing them.
type Settings struct {
	// The Discord ID of the server.
	ServerID string
	// The Discord ID of the channel where the bot will announce streams.
	AnnounceChannel StringSet
	// The Discord ID of the role that will be pinged when a stream is announced.
	AnnounceRole StringSet
	// A flag to determine if the server wants Playstation stream announcements.
	Playstation BoolSet
	// A flag to determine if the server wants Xbox stream announcements.
	Xbox BoolSet
	// A flag to determine if the server wants Nintendo stream announcements.
	Nintendo BoolSet
	// A flag to determine if the server wants PC stream announcements.
	PC BoolSet
	// A flag to determine if the server wants VR stream announcements.
	VR BoolSet
	// A flag to determine if the server settings should be reset to default values.
	Reset bool
}

// StringSet is a struct that contains a string value and a boolean flag to determine
// if the value has been set.
type StringSet struct {
	// The string value.
	Value string
	// A flag to determine if the value has been set.
	Set bool
}

// BoolSet is a struct that contains a boolean value and a boolean flag to determine
// if the value has been set.
type BoolSet struct {
	// The boolean value.
	Value bool
	// A flag to determine if the value has been set.
	Set bool
}

// NewSettings returns a new Settings struct with default values and the given server ID.
func NewSettings(serverID string) Settings {
	return Settings{
		ServerID:        serverID,
		AnnounceChannel: StringSet{"", false},
		AnnounceRole:    StringSet{"", false},
		Playstation:     BoolSet{false, false},
		Xbox:            BoolSet{false, false},
		Nintendo:        BoolSet{false, false},
		PC:              BoolSet{false, false},
		VR:              BoolSet{false, false},
		Reset:           false,
	}
}

// Set will write the values of the Settings struct to the server_settings table of the
// database. If the server is not in the table, it will insert a new row. If the server
// is in the table, it will update the row. If the server is not in the servers table,
// it will first insert a new record in that table.
func (s *Settings) Set() error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()
	logs.LogInfo("   DB", "applying settings", false, "server", s.ServerID, "settings", s)

	if !CheckSettings(s.ServerID) {
		inServerTable, err := CheckServerID(s.ServerID)
		if err != nil {
			return err
		}
		if !inServerTable {
			_, execErr := db.Exec(`INSERT INTO servers
									(server_id,
									server_name,
									owner_id,
									date_joined)
								VALUES (?, ?, ?, ?)`,
				s.ServerID,
				"",
				"",
				time.Now().UTC().Format("2006-01-02"),
				0)
			if execErr != nil {
				return execErr
			}
		}

		_, execErr := db.Exec(`INSERT INTO server_settings
									(server_id,
									announce_channel,
									announce_role,
									playstation,
									xbox,
									nintendo,
									pc,
									vr)
								VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			s.ServerID,
			s.AnnounceChannel.Value,
			s.AnnounceRole.Value,
			s.Playstation.Value,
			s.Xbox.Value,
			s.Nintendo.Value,
			s.PC.Value,
			s.VR.Value)

		if execErr != nil {
			return execErr
		}

	} else {
		_, execErr := db.Exec(`UPDATE server_settings
								SET announce_channel = ?,
									announce_role = ?,
									playstation = ?,
									xbox = ?,
									nintendo = ?,
									pc = ?,
									vr = ?
								WHERE server_id = ?`,
			s.AnnounceChannel.Value,
			s.AnnounceRole.Value,
			s.Playstation.Value,
			s.Xbox.Value,
			s.Nintendo.Value,
			s.PC.Value,
			s.VR.Value,
			s.ServerID)

		if execErr != nil {
			return execErr
		}
	}
	return nil
}

// Get populates the Settings struct with information from the server_settings table in
// the database. It uses the server ID from the struct to query the database.
func (s *Settings) Get(serverID string) error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()
	if !CheckSettings(serverID) {
		if s.Set() != nil {
			return openErr
		}
	}
	row := db.QueryRow(`SELECT server_id,
							announce_channel,
							announce_role,
							playstation,
							xbox,
							nintendo,
							pc,
							vr
						FROM server_settings
						WHERE server_id = ?`,
		serverID)

	scanErr := row.Scan(&s.ServerID,
		&s.AnnounceChannel.Value,
		&s.AnnounceRole.Value,
		&s.Playstation.Value,
		&s.Xbox.Value,
		&s.Nintendo.Value,
		&s.PC.Value,
		&s.VR.Value)

	if scanErr != nil {
		return scanErr
	}
	return nil
}

// GetPlatformServerIDs returns a list of server IDs that have the given platform set
// to true in the servers table.
func GetPlatformServerIDs(platform string) ([]string, error) {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return nil, openErr
	}
	defer db.Close()

	platform = strings.ToLower(platform)
	logs.Log.Info.WithPrefix("   DB").Info("getting server IDs for",
		"platform", platform)

	query := fmt.Sprintf(`SELECT server_id
							FROM server_settings
							WHERE %s = 1`, platform)

	rows, queryErr := db.Query(query)

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
		logs.Log.Info.Error(err)
	}
	return serverIDs, nil
}

// Merge will merge the values of the given settings struct into the settings struct
// calling the method. If a value in the given settings struct is set, it will overwrite
// the value in the calling struct.
func (s *Settings) Merge(t Settings) {
	if t.AnnounceChannel.Set {
		s.AnnounceChannel = t.AnnounceChannel
	}
	if t.AnnounceRole.Set {
		s.AnnounceRole = t.AnnounceRole
	}
	if t.Playstation.Set {
		s.Playstation = t.Playstation
	}
	if t.Xbox.Set {
		s.Xbox = t.Xbox
	}
	if t.Nintendo.Set {
		s.Nintendo = t.Nintendo
	}
	if t.PC.Set {
		s.PC = t.PC
	}
	if t.VR.Set {
		s.VR = t.VR
	}
}

// checkOptions checks if the given server ID exists in the servers table of the
// database. Returns true if the server ID exists.
func CheckSettings(serverID string) bool {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		logs.LogError("   DB", "error opening database", "err", openErr)
		return false
	}
	defer db.Close()

	rows := db.QueryRow(`SELECT server_id
						FROM server_settings
						WHERE server_id = ?`,
		serverID)

	getErr := rows.Scan(&serverID)
	return getErr == nil
}
