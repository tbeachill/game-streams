package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/utils"
)

// GetAllServerIDs returns a list of all server IDs from the servers table
func GetAllServerIDs() ([]string, error) {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return nil, openErr
	}
	defer db.Close()

	rows, queryErr := db.Query(`SELECT server_id
								FROM servers`)

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
		serverIDs = append(serverIDs, serverID)
	}
	return serverIDs, nil
}

// CheckServerID checks if the given server ID exists in the servers table. Returns
// true if the server ID exists, false if it does not.
func CheckServerID(serverID string) (bool, error) {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return false, openErr
	}
	defer db.Close()

	row := db.QueryRow(`SELECT server_id
						FROM servers
						WHERE server_id = ?`,
		serverID)

	var checkID string
	scanErr := row.Scan(&checkID)
	if scanErr == sql.ErrNoRows {
		return false, nil
	} else if scanErr != nil {
		return false, scanErr
	}
	return true, nil
}

// GetPlatformServerIDs returns a list of server IDs that have the given platform set
// to true in the servers table.
func GetPlatformServerIDs(platform string) ([]string, error) {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return nil, openErr
	}
	defer db.Close()

	platform = strings.ToLower(platform)
	utils.Log.Info.WithPrefix(" DB").Info("getting server IDs for",
		"platform", platform)

	rows, queryErr := db.Query(`SELECT server_id
								FROM servers
								WHERE ? = ?`,
		platform,
		"1")

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
		utils.Log.Info.Error(err)
	}
	return serverIDs, nil
}
