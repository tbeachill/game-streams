package db

import (
	"database/sql"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tidwall/gjson"

	"gamestreambot/utils"
)

// TODO: add options table to db - platforms, channels, etc
// TODO: add a way of deleting by specifying a list of IDs

type Stream struct {
	ID          int
	Name        string
	Platform    string
	Date        string
	Time        string
	Description string
	URL         string
}

type Streams struct {
	Streams []Stream
}

// create the db and the streams table
func CreateDB() error {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	sqlStmt := `
	create table if not exists streams (id integer not null primary key, name text, platform text, date text, time text, description text, url text);
	`

	_, tableErr := db.Exec(sqlStmt)
	if tableErr != nil {
		return tableErr
	}
	return nil
}

// get all future streams from the db and return as an s.Streams struct
func GetUpcomingStreams() (Streams, error) {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return Streams{}, openErr
	}
	defer db.Close()

	rows, queryErr := db.Query("select name, platform, date, time, description, url from streams where date >= date('now') order by date, time limit 15")
	if queryErr != nil {
		return Streams{}, queryErr
	}
	defer rows.Close()

	var streamList Streams
	for rows.Next() {
		var stream Stream
		scanErr := rows.Scan(&stream.Name, &stream.Platform, &stream.Date, &stream.Time, &stream.Description, &stream.URL)
		if scanErr != nil {
			return Streams{}, scanErr
		}
		streamList.Streams = append(streamList.Streams, stream)
	}
	return streamList, nil
}

func GetTodaysStreams() (Streams, error) {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return Streams{}, openErr
	}
	defer db.Close()

	rows, queryErr := db.Query("select name, platform, date, time, description, url from streams where date = date('now') and time >= time('now') order by time")
	if queryErr != nil {
		return Streams{}, queryErr
	}
	defer rows.Close()

	var streamList Streams
	for rows.Next() {
		var stream Stream
		scanErr := rows.Scan(&stream.Name, &stream.Platform, &stream.Date, &stream.Time, &stream.Description, &stream.URL)
		if scanErr != nil {
			return Streams{}, scanErr
		}
		streamList.Streams = append(streamList.Streams, stream)
	}
	return streamList, nil
}

// runner function to update the streams in the db, will return early if there are any errors
// or there are no updates
func UpdateStreams() error {
	var c utils.Config
	if _, tomlErr := toml.DecodeFile(utils.ConfigFile, &c); tomlErr != nil {
		return tomlErr
	}

	commitTime, timeErr := getUpdateTime(c)
	if timeErr != nil {
		return timeErr
	}

	// if there is no date in the config file, skip the comparison
	if c.LastUpdate != "" {
		updated, updateErr := compareLastUpdates(c, commitTime)
		if updateErr != nil {
			return updateErr
		}
		if !updated {
			utils.Logger.WithPrefix("UPDAT").Info("no new streams found")
			return nil
		}
	}
	utils.Logger.WithPrefix("UPDAT").Info("found new version of toml")

	newStreamList, parseErr := parseToml(c, commitTime)
	if parseErr != nil {
		return parseErr
	}

	if dateErr := formatDate(&newStreamList); dateErr != nil {
		return dateErr
	}

	updateCount, updateErr := updateRow(&newStreamList)
	if updateErr != nil {
		return updateErr
	}

	noDupList, dupErr := checkForDuplicates(newStreamList)
	if dupErr != nil {
		return dupErr
	}
	if len(noDupList.Streams) == 0 {
		utils.Logger.WithPrefix("UPDAT").Info("no new streams found")
		return nil
	}

	if insertErr := insertStreams(noDupList, commitTime, c); insertErr != nil {
		return insertErr
	}
	addedCount := len(noDupList.Streams) - updateCount
	utils.Logger.WithPrefix("UPDAT").Infof("added %d new stream%s to database and updated %d stream%s\n", addedCount, utils.Pluralise(addedCount), updateCount, utils.Pluralise(updateCount))

	if lastErr := changeLastUpdate(c, commitTime); lastErr != nil {
		return lastErr
	}
	return nil
}

// get last commit time from github
func getUpdateTime(c utils.Config) (time.Time, error) {
	response, httpErr := http.Get(c.APIURL)
	if httpErr != nil {
		return time.Time{}, httpErr
	}
	defer response.Body.Close()

	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return time.Time{}, readErr
	}
	filename := gjson.Get(string(body), "files.#.filename")
	if !strings.Contains(filename.String(), "streams.toml") {
		return time.Time{}, nil
	}
	dt := gjson.Get(string(body), "commit.author.date")
	commitTime, cTimeErr := time.Parse(time.RFC3339, dt.String())
	if cTimeErr != nil {
		return time.Time{}, cTimeErr
	}
	return commitTime, nil
}

// check if streams.toml has been updated
func compareLastUpdates(c utils.Config, commitTime time.Time) (bool, error) {
	dbTime, err := time.Parse(time.RFC3339, c.LastUpdate)
	if err != nil {
		return false, err
	}
	if commitTime.After(dbTime) || c.LastUpdate == "" {
		return true, nil
	}
	return false, nil
}

// parse streams.toml and return as an s.Streams struct
func parseToml(c utils.Config, commitTime time.Time) (Streams, error) {
	response, httpErr := http.Get(c.StreamURL)
	if httpErr != nil {
		return Streams{}, httpErr
	}
	defer response.Body.Close()

	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return Streams{}, readErr
	}
	var streamList Streams
	_, tomlErr := toml.Decode(string(body), &streamList)
	if tomlErr != nil {
		return Streams{}, tomlErr
	}
	return streamList, nil
}

// format the date and time from the toml file
func formatDate(streamList *Streams) error {
	for i, s := range streamList.Streams {
		d, err := utils.ParseTomlDate(s.Date)
		if err != nil {
			return err
		}
		streamList.Streams[i].Date = d
	}
	return nil
}

// update an existing stream in the db
func updateRow(streamList *Streams) (int, error) {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return 0, openErr
	}
	defer db.Close()

	sqlStmt := `
	update streams set name = ?, platform = ?, date = ?, time = ?, description = ?, url = ? where id = ?
	`

	var updateCount int
	for i, stream := range streamList.Streams {
		if stream.ID != 0 {
			_, updateErr := db.Exec(sqlStmt, stream.Name, stream.Platform, stream.Date, stream.Time, stream.Description, stream.URL, stream.ID)
			if updateErr != nil {
				return 0, updateErr
			}
			streamList.Streams[i] = Stream{}
			updateCount++
		}
	}
	return updateCount, nil
}

// check for duplicates in the db, if a stream in the list is found in the db, goto the next row
// if none match, add the stream to a new list
func checkForDuplicates(streamList Streams) (Streams, error) {
	rowNumber, countErr := countRows()
	if countErr != nil {
		return Streams{}, countErr
	}
	if rowNumber == 0 {
		return streamList, nil
	}

	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return Streams{}, openErr
	}
	defer db.Close()

	sqlStmt := `
	select name, platform, date, time from streams
	`

	rows, queryErr := db.Query(sqlStmt)
	if queryErr != nil {
		return Streams{}, queryErr
	}
	defer rows.Close()

	var checkedList Streams
OUTER:
	for _, s := range streamList.Streams {
		for rows.Next() {
			var stream Stream
			scanErr := rows.Scan(&stream.Name, &stream.Platform, &stream.Date, &stream.Time)
			if scanErr != nil {
				return Streams{}, scanErr
			}
			if s.Name == stream.Name && s.Platform == stream.Platform && s.Date == stream.Date && s.Time == stream.Time {
				continue OUTER
			}
		}
		checkedList.Streams = append(checkedList.Streams, s)
	}
	return checkedList, nil
}

// count the number of rows in the streams table
func countRows() (int, error) {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return 0, openErr
	}
	defer db.Close()

	sqlStmt := `
	select count(*) from streams
	`

	var count int
	row := db.QueryRow(sqlStmt)
	scanErr := row.Scan(&count)
	if scanErr != nil {
		return 0, scanErr
	}
	return count, nil
}

// update db with new streams
func insertStreams(streamList Streams, commitTime time.Time, c utils.Config) error {
	db, sqlErr := sql.Open("sqlite3", utils.DBFile)
	if sqlErr != nil {
		return sqlErr
	}
	defer db.Close()

	sqlStmt := `
	insert into streams (name, platform, date, time, description, url) values (?, ?, ?, ?, ?, ?)
	`

	for _, stream := range streamList.Streams {
		if stream.Name == "" {
			continue
		}
		_, insertErr := db.Exec(sqlStmt, stream.Name, stream.Platform, stream.Date, stream.Time, stream.Description, stream.URL)
		if insertErr != nil {
			return insertErr
		}
	}
	return nil
}

// change last update time in config.toml
func changeLastUpdate(c utils.Config, commitTime time.Time) error {
	f, fileErr := os.Create(utils.ConfigFile)
	if fileErr != nil {
		return fileErr
	}
	encErr := toml.NewEncoder(f).Encode(utils.Config{CommandURL: c.CommandURL, StreamURL: c.StreamURL, APIURL: c.APIURL, LastUpdate: commitTime.Format("2006-01-02T15:04:05Z07:00")})
	if encErr != nil {
		return encErr
	}
	closeErr := f.Close()
	if closeErr != nil {
		return closeErr
	}
	return nil
}
