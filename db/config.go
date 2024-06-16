package db

import (
	"database/sql"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"gamestreambot/utils"
)

type Config struct {
	ID         int
	StreamURL  string
	APIURL     string
	LastUpdate string
	CommitTime time.Time
}

// get the values of the config struct from the db or set default values if none are found
func (c *Config) Get() error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}

	sqlStmt := `
		select * from config where id = 1
	`
	defer db.Close()

	row := db.QueryRow(sqlStmt)
	scanErr := row.Scan(&c.ID, &c.StreamURL, &c.APIURL, &c.LastUpdate)
	if scanErr == sql.ErrNoRows {
		utils.Log.Info.WithPrefix(" MAIN").Info("No config found, setting default")
		if defaultErr := c.SetDefault(); defaultErr != nil {
			return defaultErr
		}
	} else if scanErr != nil {
		return scanErr
	}
	return nil
}

// set the default values for the config struct and write to the db
func (c *Config) SetDefault() error {
	utils.Log.Info.WithPrefix(" MAIN").Info("setting default config")
	c.StreamURL = os.Getenv("STREAM_URL")
	c.APIURL = os.Getenv("API_URL")
	c.LastUpdate = ""

	sqlStmt := `
		insert into config (id, stream_url, api_url, last_updated) values (1, ?, ?, ?)
	`
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(sqlStmt, c.StreamURL, c.APIURL, "")
	if execErr != nil {
		return execErr
	}

	return nil
}

// write the values of the config struct to the db
func (c *Config) Set() error {
	utils.Log.Info.WithPrefix(" MAIN").Info("updating config")
	sqlStmt := `
		update config set stream_url = ?, api_url = ?, last_updated = ? where id = 1
	`
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()
	c.LastUpdate = c.CommitTime.Format("2006-01-02T15:04:05Z07:00")

	_, execErr := db.Exec(sqlStmt, c.StreamURL, c.APIURL, c.LastUpdate)
	if execErr != nil {
		return execErr
	}
	return nil
}

// get the last updated time from the db and compare to the last commit time from the api
// return true if the commit time is newer than the last updated time
func (c *Config) Check() (bool, error) {
	if c.LastUpdate == "" {
		return true, nil
	}
	utils.Log.Info.WithPrefix("UPDAT").Info("getting last update time")
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

	utils.Log.Info.WithPrefix("UPDAT").Info("comparing update times")
	dbTime, parseErr := time.Parse(time.RFC3339, c.LastUpdate)
	if parseErr != nil {
		return false, parseErr
	}
	if commitTime.After(dbTime) || c.LastUpdate == "" {
		return true, nil
	}
	return false, nil
}
