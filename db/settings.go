package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/utils"
)

// Options is a struct that contains the options for a server. It contains the server ID, the channel to announce
// streams, the role to announce streams, and flags for each platform to announce streams for. It also contains a flag
// to reset the options for a server.
type Options struct {
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

// StringSet is a struct that contains a string value and a boolean flag to determine if the value has been set.
type StringSet struct {
	Value string
	Set   bool
}

// BoolSet is a struct that contains a boolean value and a boolean flag to determine if the value has been set.
type BoolSet struct {
	Value bool
	Set   bool
}

// NewOptions returns a new Options struct with default values and the given server ID.
func NewOptions(serverID string) Options {
	return Options{
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

// Set will write the values of the options struct to the servers table of the database. If the server is not in the
// table, it will insert a new row. If the server is in the table, it will update the row.
func (o *Options) Set() error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()
	utils.Log.Info.Info("setting options", "server", o.ServerID, "options", o)

	if !checkOptions(o.ServerID) {
		_, execErr := db.Exec("insert into servers (server_id, announce_channel, announce_role, playstation, xbox, nintendo, pc, vr) values (?, ?, ?, ?, ?, ?, ?, ?)", o.ServerID, o.AnnounceChannel.Value, o.AnnounceRole.Value, o.Playstation.Value, o.Xbox.Value, o.Nintendo.Value, o.PC.Value, o.VR.Value)
		if execErr != nil {
			return execErr
		}
	} else {
		_, execErr := db.Exec("update servers set announce_channel = ?, announce_role = ?, playstation = ?, xbox = ?, nintendo = ?, pc = ?, vr = ? where server_id = ?", o.AnnounceChannel.Value, o.AnnounceRole.Value, o.Playstation.Value, o.Xbox.Value, o.Nintendo.Value, o.PC.Value, o.VR.Value, o.ServerID)
		if execErr != nil {
			return execErr
		}
	}
	return nil
}

// Get will get the settings for a server from the servers table of the database and write them to the options struct.
// If the server is not in the table, it will set the default values for the options struct and write them to the
// database.
func (o *Options) Get(serverID string) error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()
	if !checkOptions(serverID) {
		if o.Set() != nil {
			return openErr
		}
	}
	row := db.QueryRow("select server_id, announce_channel, announce_role, playstation, xbox, nintendo, pc, vr from servers where server_id = ?", serverID)
	scanErr := row.Scan(&o.ServerID, &o.AnnounceChannel.Value, &o.AnnounceRole.Value, &o.Playstation.Value, &o.Xbox.Value, &o.Nintendo.Value, &o.PC.Value, &o.VR.Value)
	if scanErr != nil {
		return scanErr
	}
	return nil
}

// Merge will merge the values of the given options struct into the options struct calling the method. If a value is set
// in the given options struct, it will overwrite the value in the calling options struct.
func (o *Options) Merge(p Options) {
	if p.AnnounceChannel.Set {
		o.AnnounceChannel = p.AnnounceChannel
	}
	if p.AnnounceRole.Set {
		o.AnnounceRole = p.AnnounceRole
	}
	if p.Playstation.Set {
		o.Playstation = p.Playstation
	}
	if p.Xbox.Set {
		o.Xbox = p.Xbox
	}
	if p.Nintendo.Set {
		o.Nintendo = p.Nintendo
	}
	if p.PC.Set {
		o.PC = p.PC
	}
	if p.VR.Set {
		o.VR = p.VR
	}
}

// checkOptions checks if the given server ID exists in the servers table of the database. Returns true if the server
// ID exists, false if it does not.
func checkOptions(serverID string) bool {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		utils.Log.ErrorWarn.Error("error opening db", "error", openErr)
		return false
	}
	defer db.Close()

	rows := db.QueryRow("select server_id from servers where server_id = ?", serverID)
	getErr := rows.Scan(&serverID)
	return getErr == nil
}

// RemoveOptions will remove the server from the servers table of the database.
func RemoveOptions(serverID string) error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	utils.Log.Info.WithPrefix("STATS").Info("removing from options table", "server", serverID)
	_, execErr := db.Exec("delete from servers where server_id = ?", serverID)
	if execErr != nil {
		return execErr
	}
	return nil
}
