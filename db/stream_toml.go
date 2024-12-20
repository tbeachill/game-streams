/*
stream_toml.go contains the StreamTOML struct and methods for interacting with the
stream_toml table in the database. This table contains information about the streams.toml
file, including the last time it was updated and the time of the last commit to the file.
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

// StreamToml holds information about the streams.toml file.
var StreamToml StreamTOML

// StreamTOML represents a row in the stream_toml table of the database.
type StreamTOML struct {
	// The ID of the row.
	ID int
	// The time the streams table was last updated with the streams.toml file.
	LastUpdate string
	// The time of the last commit to the streams.toml file.
	CommitTime time.Time
}

// Get retrieves the stream_toml values from the database and stores them in the struct.
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
		logs.LogInfo("   DB", "No stream_toml values found, setting default", false)
		if defaultErr := t.SetDefault(); defaultErr != nil {
			return defaultErr
		}
	} else if scanErr != nil {
		return scanErr
	}
	return nil
}

// SetDefault sets the default values for the stream_toml table in the database.
func (t *StreamTOML) SetDefault() error {
	logs.LogInfo("   DB", "Setting default stream_toml values", false)

	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`INSERT INTO stream_toml
								(id,
								last_updated)
							VALUES (1, "")`)

	if execErr != nil {
		return execErr
	}

	return nil
}

// Set writes the current values of the struct to the stream_toml table in the database.
func (t *StreamTOML) Set() error {
	logs.LogInfo("   DB", "Updating stream_toml values", false)

	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()
	t.LastUpdate = t.CommitTime.Format("2006-01-02T15:04:05Z07:00")

	_, execErr := db.Exec(`UPDATE stream_toml
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
