package db

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"gamestreams/config"
	"gamestreams/logs"
)

// Settings is a struct that contains the options for a server. It contains the
// server ID, the channel to announce streams, the role to announce streams, and flags
// for each platform to announce streams for. It also contains a flag to reset the
// options for a server.
type Settings struct {
	ServerID        string
	AnnounceChannel StringSet
	AnnounceRole    StringSet
	Playstation     BoolSet
	Xbox            BoolSet
	Nintendo        BoolSet
	PC              BoolSet
	VR              BoolSet
	Reset           bool
}

// StringSet is a struct that contains a string value and a boolean flag to determine
// if the value has been set.
type StringSet struct {
	Value string
	Set   bool
}

// BoolSet is a struct that contains a boolean value and a boolean flag to determine
// if the value has been set.
type BoolSet struct {
	Value bool
	Set   bool
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
// is in the table, it will update the row.
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

// Get will get the settings for a server from the servers table of the database and
// write them to the options struct. If the server is not in the table, it will set the
// default values for the options struct and write them to the database.
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

// Merge will merge the values of the given options struct into the options struct
// calling the method. If a value is set in the given options struct, it will overwrite
// the value in the calling options struct.
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
// database. Returns true if the server ID exists, false if it does not.
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

// RemoveServerSettings removes the settings for the given server ID from the
// server_settings table of the database.
func RemoveServerSettings(serverID string) error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`DELETE FROM server_settings
							WHERE server_id = ?`,
		serverID)
	if execErr != nil {
		return execErr
	}
	return nil
}
