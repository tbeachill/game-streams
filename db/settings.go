package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/utils"
)

type Options struct {
	ServerID        string
	AnnounceChannel StringSet
	AnnounceRole    StringSet
	Playstation     BoolSet
	Xbox            BoolSet
	Nintendo        BoolSet
	PC              BoolSet
	Reset           bool
}
type StringSet struct {
	Value string
	Set   bool
}
type BoolSet struct {
	Value bool
	Set   bool
}

func NewOptions(serverID string) Options {
	return Options{
		ServerID:        serverID,
		AnnounceChannel: StringSet{"", false},
		AnnounceRole:    StringSet{"", false},
		Playstation:     BoolSet{false, false},
		Xbox:            BoolSet{false, false},
		Nintendo:        BoolSet{false, false},
		PC:              BoolSet{false, false},
		Reset:           false,
	}
}

// set the options for a server
func (o *Options) Set() error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()
	checkOptions(o.ServerID)
	utils.Log.Info.Info("setting options", "server", o.ServerID, "options", o)

	_, execErr := db.Exec("update settings set announce_channel = ?, announce_role = ?, playstation = ?, xbox = ?, nintendo = ?, pc = ? where server_id = ?", o.AnnounceChannel.Value, o.AnnounceRole.Value, o.Playstation.Value, o.Xbox.Value, o.Nintendo.Value, o.PC.Value, o.ServerID)
	if execErr != nil {
		return execErr
	}
	return nil
}

// get the options for a server and set the options struct
func (o *Options) Get(serverID string) error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()
	checkOptions(serverID)

	row := db.QueryRow("select * from settings where server_id = ?", serverID)
	scanErr := row.Scan(&o.ServerID, &o.AnnounceChannel.Value, &o.AnnounceRole.Value, &o.Playstation.Value, &o.Xbox.Value, &o.Nintendo.Value, &o.PC.Value)
	if scanErr != nil {
		return scanErr
	}
	return nil
}

// merge the options from a new options struct with the current options struct
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
}

// check if a server is in the settings table, if not add it with default options
func checkOptions(serverID string) error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	rows := db.QueryRow("select server_id from settings where server_id = ?", serverID)
	getErr := rows.Scan(&serverID)
	if getErr != nil {
		o := NewOptions(serverID)
		if setErr := o.Set(); setErr != nil {
			return setErr
		}
	}
	return nil
}

// remove a server from the settings table
func RemoveOptions(serverID string) error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	utils.Log.Info.WithPrefix("STATS").Info("removing from options table", "server", serverID)
	_, execErr := db.Exec("delete from settings where server_id = ?", serverID)
	if execErr != nil {
		return execErr
	}
	return nil
}
