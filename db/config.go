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

var Conf Config

// Config is a struct that holds the configuration values for the bot.
type Config struct {
	ID             int
	StreamsTOMLURL string
	APIURL         string
	LastUpdate     string
	CommitTime     time.Time
}

// Get gets the configuration values from the database and sets them in the Config
// struct.
func (c *Config) Get() error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	row := db.QueryRow(`SELECT *
						FROM config
						WHERE id = 1`)

	scanErr := row.Scan(&c.ID, &c.StreamsTOMLURL, &c.APIURL, &c.LastUpdate)
	if scanErr == sql.ErrNoRows {
		logs.LogInfo("   DB", "No config found, setting default", false)
		if defaultErr := c.SetDefault(); defaultErr != nil {
			return defaultErr
		}
	} else if scanErr != nil {
		return scanErr
	}
	return nil
}

// SetDefault sets the default values for the config struct from the environment
// variables and writes them to the database.
func (c *Config) SetDefault() error {
	logs.LogInfo("   DB", "Setting default config", false)
	c.StreamsTOMLURL = config.Values.Github.StreamsTOMLURL
	c.APIURL = config.Values.Github.APIURL
	c.LastUpdate = ""

	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`INSERT INTO config
								(id,
								toml_url,
								api_url,
								last_updated)
							VALUES (1, ?, ?, ?)`,
		c.StreamsTOMLURL,
		c.APIURL,
		"")

	if execErr != nil {
		return execErr
	}

	return nil
}

// Set writes the current values of the Config struct to the database.
func (c *Config) Set() error {
	logs.LogInfo("   DB", "Updating config", false)

	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()
	c.LastUpdate = c.CommitTime.Format("2006-01-02T15:04:05Z07:00")

	_, execErr := db.Exec(`UPDATE config
							SET toml_url = ?,
								api_url = ?,
								last_updated = ?
							WHERE id = 1`,
		c.StreamsTOMLURL,
		c.APIURL,
		c.LastUpdate)

	if execErr != nil {
		return execErr
	}
	return nil
}

// Check checks whether the streams.toml file has been updated. It gets the time of the
// last commit to the streams.toml in the flat-files repository and compares it to the
// last update time stored in the database. If the commit time is after the last update
// time, it returns true.
func (c *Config) Check() (bool, error) {
	if c.LastUpdate == "" {
		return true, nil
	}
	response, httpErr := http.Get(c.APIURL)
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
	c.CommitTime = commitTime

	dbTime, parseErr := time.Parse(time.RFC3339, c.LastUpdate)
	if parseErr != nil {
		return false, parseErr
	}
	if commitTime.After(dbTime) || c.LastUpdate == "" {
		return true, nil
	}
	return false, nil
}
