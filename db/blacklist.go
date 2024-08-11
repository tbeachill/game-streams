package db

import (
	"database/sql"
	"time"

	"gamestreams/config"
	"gamestreams/logs"
)

// Blacklist is a struct that holds the blacklist values for the bot.
type Blacklist struct {
	ID          int
	IDType      string
	DateAdded   string
	DateExpires string
	Reason      string
}

// IsBlacklisted checks if the given ID is blacklisted. Returns true if the ID is blacklisted,
// false if it is not.
func IsBlacklisted(id string, idType string) (bool, string, string) {
	logs.LogInfo(" MAIN", "checking if blacklisted", false,
		"id", id,
		"idType", idType)
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return false, "", ""
	}
	defer db.Close()

	var query string
	if idType == "" {
		query = `SELECT reason,
						date_expires
				FROM blacklist
				WHERE discord_id = ?`
	} else {
		query = `SELECT reason,
						date_expires
				FROM blacklist
				WHERE discord_id = ?
				AND id_type = ?`
	}
	row := db.QueryRow(query,
		id, idType)

	var reason, expiryDate string
	scanErr := row.Scan(&reason, &expiryDate)
	return scanErr == nil, reason, expiryDate
}

// AddToBlacklist adds the given ID to the blacklist table of the database.
func AddToBlacklist(id string, idType string, reason string, length_days int) error {
	logs.LogInfo("OWNER", "adding to blacklist table", false,
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
	logs.LogInfo("OWNER", "removing from blacklist table", false, "id", id)
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
	logs.LogInfo("OWNER", "getting blacklist", false)
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
							WHERE date_expires <= date('now')`)
	return execErr
}
