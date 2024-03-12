package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/utils"
)

// get a list of all server IDs from the settings table
func GetServerIDs() ([]string, error) {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return nil, openErr
	}
	defer db.Close()

	rows, queryErr := db.Query("select server_id from settings")
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

// get a list of all server IDs from the settings table where the platform is true
func GetPlatformServerIDs(platform string) ([]string, error) {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return nil, openErr
	}
	defer db.Close()

	platform = strings.ToLower(platform)
	utils.Log.Info.WithPrefix(" DB").Info("getting server IDs for", "platform", platform)
	query := fmt.Sprintf("select server_id from settings where %s = %s", platform, "1")
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
