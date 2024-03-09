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
	Awards          BoolSet
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

// add a server to the settings table with default options
func SetDefaultOptions(serverID string) error {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	utils.Logger.WithPrefix("STATS").Info("setting default options", "server", serverID)
	_, execErr := db.Exec("insert into settings (server_id, announce_channel, announce_role, playstation, xbox, nintendo, pc, awards) values (?, ?, ?, ?, ?, ?, ?, ?)", serverID, "", "", 0, 0, 0, 0, 0)
	if execErr != nil {
		return execErr
	}
	return nil
}

// reset the options for a server to default
func ResetOptions(serverID string) error {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return openErr
	}
	defer db.Close()
	utils.Logger.WithPrefix(" CMND").Info("resetting options", "server", serverID)
	_, execErr := db.Exec("update settings set announce_channel = ?, announce_role = ?, playstation = ?, xbox = ?, nintendo = ?, pc = ?, awards = ? where server_id = ?", "", "", 0, 0, 0, 0, 0, serverID)
	if execErr != nil {
		return execErr
	}
	return nil
}

// remove a server from the settings table
func RemoveOptions(serverID string) error {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	utils.Logger.WithPrefix("STATS").Info("removing from options table", "server", serverID)
	_, execErr := db.Exec("delete from settings where server_id = ?", serverID)
	if execErr != nil {
		return execErr
	}
	return nil
}

// set the options for a server from an options struct
func SetOptions(options *Options) error {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return openErr
	}
	defer db.Close()
	checkOptions(options.ServerID)
	utils.Logger.Info("setting options", "server", options.ServerID, "options", options)

	_, execErr := db.Exec("update settings set announce_channel = ?, announce_role = ?, playstation = ?, xbox = ?, nintendo = ?, pc = ?, awards = ? where server_id = ?", options.AnnounceChannel.Value, options.AnnounceRole.Value, options.Playstation.Value, options.Xbox.Value, options.Nintendo.Value, options.PC.Value, options.Awards.Value, options.ServerID)
	if execErr != nil {
		return execErr
	}
	return nil
}

// get the options for a server and return as an options struct
func GetOptions(serverID string) (Options, error) {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return Options{}, openErr
	}
	defer db.Close()
	checkOptions(serverID)

	var options Options
	row := db.QueryRow("select * from settings where server_id = ?", serverID)
	scanErr := row.Scan(&options.ServerID, &options.AnnounceChannel.Value, &options.AnnounceRole.Value, &options.Playstation.Value, &options.Xbox.Value, &options.Nintendo.Value, &options.PC.Value, &options.Awards.Value)
	if scanErr != nil {
		return Options{}, scanErr
	}
	return options, nil
}

// check if a server is in the settings table, if not add it with default options
func checkOptions(serverID string) error {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	rows := db.QueryRow("select server_id from settings where server_id = ?", serverID)
	getErr := rows.Scan(&serverID)
	if getErr != nil {
		setErr := SetDefaultOptions(serverID)
		if setErr != nil {
			return setErr
		}
	}
	return nil
}

// merge the options from a new options struct with the current options struct
func MergeOptions(serverID string, new *Options) *Options {
	current, getErr := GetOptions(serverID)
	if getErr != nil {
		utils.EWLogger.WithPrefix(" CMND").Error("error getting options", "server", serverID, "err", getErr)
		return &Options{}
	}
	if new.AnnounceChannel.Set {
		current.AnnounceChannel = new.AnnounceChannel
	}
	if new.AnnounceRole.Set {
		current.AnnounceRole = new.AnnounceRole
	}
	if new.Playstation.Set {
		current.Playstation = new.Playstation
	}
	if new.Xbox.Set {
		current.Xbox = new.Xbox
	}
	if new.Nintendo.Set {
		current.Nintendo = new.Nintendo
	}
	if new.PC.Set {
		current.PC = new.PC
	}
	if new.Awards.Value {
		current.Awards = new.Awards
	}
	return &current
}
