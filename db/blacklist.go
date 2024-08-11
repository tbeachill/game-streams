package db

import (
	"database/sql"
	"time"

	"gamestreams/config"
	"gamestreams/logs"
)

// Blacklist is a struct that holds the blacklist values for the bot.
type Blacklist struct {
	ID           int
	IDType       string
	DateAdded    string
	DateExpires  string
	Reason       string
	LastMessaged string
}

// IsBlacklisted checks if the given ID is blacklisted. Returns true if the ID is blacklisted,
// false if it is not.
func IsBlacklisted(id string, idType string) (bool, Blacklist) {
	logs.LogInfo("   DB", "checking if blacklisted", false,
		"id", id,
		"idType", idType)
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return false, Blacklist{}
	}
	defer db.Close()

	var query string
	var row *sql.Row
	if idType == "" {
		query = `SELECT id,
						id_type,
						date_added,
						date_expires,
						reason,
						last_messaged
				FROM blacklist
				WHERE discord_id = ?`
		row = db.QueryRow(query, id)
	} else {
		query = `SELECT id,
					id_type,
					date_added,
					date_expires,
					reason,
					last_messaged
				FROM blacklist
				WHERE discord_id = ?
				AND id_type = ?`
		row = db.QueryRow(query, id, idType)
	}
	var b Blacklist
	scanErr := row.Scan(&b.ID, &b.IDType, &b.DateAdded, &b.DateExpires, &b.Reason, &b.LastMessaged)
	if scanErr != nil {
		logs.LogInfo("   DB", "not blacklisted", false,
			"id", id,
			"idType", idType)
		return false, Blacklist{}
	}
	logs.LogInfo("   DB", "blacklisted", false,
		"id", id,
		"idType", idType,
		"reason", b.Reason)
	return true, b
}

// AddToBlacklist adds the given ID to the blacklist table of the database.
func AddToBlacklist(id string, idType string, reason string, length_days int) error {
	logs.LogInfo("   DB", "adding to blacklist table", false,
		"id", id,
		"idType", idType,
		"days", length_days,
		"reason", reason)

	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	currentDate := time.Now().UTC().Format("2006-01-02")
	expiryDate := time.Now().AddDate(0, 0, length_days).UTC().Format("2006-01-02")
	_, execErr := db.Exec(`INSERT INTO blacklist
								(discord_id,
								id_type,
								date_added,
								date_expires,
								reason)
							VALUES (?, ?, ?, ?, ?)`,
		id, idType, currentDate, expiryDate, reason)
	return execErr
}

// RemoveFromBlacklist removes the given ID from the blacklist table of the database.
func RemoveFromBlacklist(id string) error {
	logs.LogInfo("   DB", "removing from blacklist table", false, "id", id)
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`DELETE FROM blacklist
							WHERE discord_id = ?`,
		id)
	return execErr
}

// GetBlacklist returns a list of all blacklisted IDs from the blacklist table of the database.
func GetBlacklist() ([]Blacklist, error) {
	logs.LogInfo("   DB", "getting blacklist", false)
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return nil, openErr
	}
	defer db.Close()

	rows, queryErr := db.Query(`SELECT discord_id,
									id_type,
									date_added,
									date_expires,
									reason
								FROM blacklist`)
	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()

	var blacklist []Blacklist
	for rows.Next() {
		var b Blacklist
		scanErr := rows.Scan(&b.ID, &b.IDType, &b.DateAdded, &b.DateExpires, &b.Reason)
		if scanErr != nil {
			return nil, scanErr
		}
		blacklist = append(blacklist, b)
	}
	return blacklist, nil
}

// RemoveExpiredBlacklist removes all blacklisted IDs that have expired from the blacklist table
// of the database.
func RemoveExpiredBlacklist() error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`DELETE FROM blacklist
							WHERE date_expires <= DATE('now')`)
	return execErr
}

// UpdateLastMessaged updates the last_messaged field of the given ID in the blacklist table of
// the database.
func UpdateLastMessaged(id string) error {
	logs.LogInfo("   DB", "updating last messaged", false, "id", id)
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`UPDATE blacklist
							SET last_messaged = ?
							WHERE discord_id = ?`,
		time.Now().UTC().Format("2006-01-02"), id)
	return execErr
}
