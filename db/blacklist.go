package db

import (
	"database/sql"

	"gamestreambot/utils"
)

// Blacklist is a struct that holds the blacklist values for the bot.
type Blacklist struct {
	ID     int
	IDType string
	Date   string
	Reason string
}

// IsBlacklisted checks if the given ID is blacklisted. Returns true if the ID is blacklisted,
// false if it is not.
func IsBlacklisted(id string, idType string) (bool, string) {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return false, ""
	}
	defer db.Close()

	row := db.QueryRow(`SELECT reason
						FROM blacklist
						WHERE id = ?
						AND id_type = ?`,
		id, idType)

	var reason string
	scanErr := row.Scan(&reason)
	return scanErr == nil, reason
}
