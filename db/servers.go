package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/utils"
)

// get a list of all server IDs that the bot is present in from the servers table
func GetServerIDs() ([]string, error) {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return nil, openErr
	}
	defer db.Close()

	rows, queryErr := db.Query("select server_id from servers")
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

// check if a server ID is in the servers table
func CheckServerID(serverID string) (bool, error) {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return false, openErr
	}
	defer db.Close()

	row := db.QueryRow("select server_id from servers where server_id = ?", serverID)
	var checkID string
	scanErr := row.Scan(&checkID)
	if scanErr == sql.ErrNoRows {
		return false, nil
	} else if scanErr != nil {
		return false, scanErr
	}
	return true, nil
}

// get a list of all server IDs from the servers table where the given platform is true
func GetPlatformServerIDs(platform string) ([]string, error) {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return nil, openErr
	}
	defer db.Close()

	platform = strings.ToLower(platform)
	utils.Log.Info.WithPrefix(" DB").Info("getting server IDs for", "platform", platform)
	query := fmt.Sprintf("select server_id from servers where %s = %s", platform, "1")
	rows, queryErr := db.Query(query)
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
