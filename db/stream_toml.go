/*
config.go provides functions to interact with the config table of the database.
*/
package db

import (
	"database/sql"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"gamestreams/config"
	"gamestreams/logs"
)

var StreamToml StreamTOML

// Config is a struct that holds the configuration values for the bot.
type StreamTOML struct {
	ID         int
	LastUpdate string
	CommitTime time.Time
}

// Get gets the configuration values from the database and sets them in the Config
// struct.
func (t *StreamTOML) Get() error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	row := db.QueryRow(`SELECT *
						FROM stream_toml
						WHERE id = 1`)

	scanErr := row.Scan(&t.ID, &t.LastUpdate)
	if scanErr == sql.ErrNoRows {
		logs.LogInfo("   DB", "No config found, setting default", false)
		if defaultErr := t.SetDefault(); defaultErr != nil {
			return defaultErr
		}
	} else if scanErr != nil {
		return scanErr
	}
	return nil
}

// SetDefault sets the default values for the config struct from the environment
// variables and writes them to the database.
func (t *StreamTOML) SetDefault() error {
	logs.LogInfo("   DB", "Setting default config", false)

	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`INSERT INTO config
								(id,
								last_updated)
							VALUES (1, ?)`,
		"")

	if execErr != nil {
		return execErr
	}

	return nil
}

// Set writes the current values of the Config struct to the database.
func (t *StreamTOML) Set() error {
	logs.LogInfo("   DB", "Updating config", false)

	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()
	t.LastUpdate = t.CommitTime.Format("2006-01-02T15:04:05Z07:00")

	_, execErr := db.Exec(`UPDATE config
							SET last_updated = ?
							WHERE id = 1`,
		t.LastUpdate)

	if execErr != nil {
		return execErr
	}
	return nil
}

// Check checks whether the streams.toml file has been updated. It gets the time of the
// last commit to the streams.toml in the flat-files repository and compares it to the
// last update time stored in the database. If the commit time is after the last update
// time, it returns true.
func (t *StreamTOML) Check() (bool, error) {
	if t.LastUpdate == "" {
		return true, nil
	}
	response, httpErr := http.Get(config.Values.Github.APIURL)
	if httpErr != nil {
		return false, httpErr
	}
	defer response.Body.Close()

	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return false, readErr
	}
	filename := gjson.Get(string(body), "files.#.filename")
	if !strings.Contains(filename.String(), "streams.toml") {
		return false, nil
	}
	dt := gjson.Get(string(body), "commit.author.date")
	commitTime, cTimeErr := time.Parse(time.RFC3339, dt.String())
	if cTimeErr != nil {
		return false, cTimeErr
	}
	t.CommitTime = commitTime

	dbTime, parseErr := time.Parse(time.RFC3339, t.LastUpdate)
	if parseErr != nil {
		return false, parseErr
	}
	if commitTime.After(dbTime) || t.LastUpdate == "" {
		return true, nil
	}
	return false, nil
}
