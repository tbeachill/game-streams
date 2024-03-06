package db

import (
	"database/sql"

	"gamestreambot/utils"
)

func GetConfig() utils.Config {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		utils.Logger.Error(openErr)
		return utils.Config{}
	}

	sqlStmt := `
		select * from config where id = 1
	`
	defer db.Close()

	var config utils.Config
	row := db.QueryRow(sqlStmt)
	scanErr := row.Scan(&config.ID, &config.StreamURL, &config.APIURL, &config.LastUpdate)
	if scanErr == sql.ErrNoRows {
		utils.Logger.Info("No config found, setting default")
		if defaultErr := SetDefaultConfig(); defaultErr != nil {
			utils.Logger.Error(defaultErr)
		}
		return GetConfig()
	} else if scanErr != nil {
		utils.Logger.Error(openErr)
		return utils.Config{}
	}
	return config
}

func SetConfig(config utils.Config) error {
	sqlStmt := `
		update config set stream_url = ?, api_url = ?, last_updated = ? where id = 1
	`
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(sqlStmt, config.StreamURL, config.APIURL, config.LastUpdate)
	if execErr != nil {
		return execErr
	}
	return nil
}

func SetDefaultConfig() error {
	sqlStmt := `
		insert into config (stream_url, api_url, last_updated) values (?, ?, ?)
	`
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(sqlStmt, "https://raw.githubusercontent.com/tbeachill/flat-files/main/streams.toml",
		"https://api.github.com/repos/tbeachill/flat-files/commits/main",
		"")
	if execErr != nil {
		return execErr
	}
	return nil
}
