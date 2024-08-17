/*
blacklist.go provides functions to interact with the blacklist table of the database.
*/
package db

import (
	"database/sql"
	"time"

	"gamestreams/config"
	"gamestreams/logs"
)

// Blacklist is a struct that holds the values from a row in the blacklist table
// of the database.
type Blacklist struct {
	// The ID of the blacklisted user/server.
	ID int
	// The type of ID that was blacklisted (user/server).
	IDType string
	// The date the ID was added to the blacklist.
	DateAdded string
	// The date the ID will expire from the blacklist.
	DateExpires string
	// The reason the ID was blacklisted.
	Reason string
	// The date the ID was last messaged explaining the blacklist reason.
	LastMessaged string
}

// IsBlacklisted checks if the given ID is blacklisted. Returns true and the blacklist
// values if the ID is blacklisted, otherwise returns false and an empty Blacklist struct.
func IsBlacklisted(id string) (bool, Blacklist) {
	logs.LogInfo("   DB", "checking if blacklisted", false,
		"id", id)
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return false, Blacklist{}
	}
	defer db.Close()

	row := db.QueryRow(`SELECT id,
							id_type,
							date_added,
							date_expires,
							reason,
							last_messaged
						FROM blacklist
						WHERE discord_id = ?
						AND date_expires > ?`,
		id,
		time.Now().UTC().Format("2006-01-02"))

	var b Blacklist
	scanErr := row.Scan(&b.ID, &b.IDType, &b.DateAdded, &b.DateExpires, &b.Reason, &b.LastMessaged)
	if scanErr != nil {
		return false, Blacklist{}
	}
	return true, b
}

// AddToBlacklist adds the given ID to the blacklist table. The ID is
// added with the given ID type, reason, and length of time in days.
func AddToBlacklist(id string, idType string, reason string, length_days int) error {
	blacklisted, _ := IsBlacklisted(id)
	if blacklisted {
		logs.LogInfo("   DB", "ID already blacklisted", false, "id", id)
		return nil
	}
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

// RemoveFromBlacklist removes the given ID from the blacklist table.
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

// GetBlacklist returns a slice of Blacklist structs of all IDs in the blacklist.
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
								FROM blacklist
								WHERE date_expires > ?
								ORDER BY date_expires ASC`,
		time.Now().UTC().Format("2006-01-02"))
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

// UpdateLastMessaged updates the last_messaged field of the given ID in the blacklist
// table to the current date.
func UpdateLastMessaged(id string) error {
	logs.LogInfo("   DB", "updating last messaged", false, "id", id)
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	timeNow := time.Now().UTC().Format("2006-01-02")

	_, execErr := db.Exec(`UPDATE blacklist
							SET last_messaged = ?
							WHERE discord_id = ?
							AND date_expires > ?`,
		timeNow,
		id,
		timeNow)
	return execErr
}
