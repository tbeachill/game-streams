/*
blacklist.go provides functions to interact with the blacklist table of the database.
*/
package db

import (
	"database/sql"
	"math"
	//"time"

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
						AND date_expires > DATE('now')`,
		id)

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

	bCount, countErr := countBlacklistEntries(id, 2)
	if countErr != nil {
		return countErr
	}
	if bCount > 0 {
		length_days = int(math.Pow(float64(length_days), float64(bCount)))
		if length_days > 365 {
			length_days = 365
		}
	}

	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`INSERT INTO blacklist
								(discord_id,
								id_type,
								date_added,
								date_expires,
								reason)
							VALUES (?, ?, DATE('now'), DATE('now', ? || ' days'), ?)`,
		id, idType, length_days, reason)
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
								WHERE date_expires > DATE('now')
								ORDER BY date_expires ASC`)
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

// CountBlacklistEntries returns the number of entries in the blacklist table for the
// given ID.
func countBlacklistEntries(id string, num_years int) (int, error) {
	logs.LogInfo("   DB", "counting blacklist entries", false, "id", id)
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return 0, openErr
	}
	defer db.Close()

	row := db.QueryRow(`SELECT COUNT(*)
						FROM blacklist
						WHERE discord_id = ?
						AND date_expires > DATE('now', ? || ' years')`,
		id,
		-num_years)

	var count int
	scanErr := row.Scan(&count)
	if scanErr != nil {
		return 0, scanErr
	}
	return count, nil
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

	_, execErr := db.Exec(`UPDATE blacklist
							SET last_messaged = DATE('now')
							WHERE discord_id = ?
							AND date_expires > DATE('now')`,
		id)
	return execErr
}
