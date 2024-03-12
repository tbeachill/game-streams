package db

import (
	"database/sql"
	"io"
	"net/http"

	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/utils"
)

// runner function to update the streams in the db, will return early if there are any errors
// or there are no updates
func UpdateStreams() error {
	var c Config
	c.Get()
	updated, checkErr := c.Check()
	if checkErr != nil {
		return checkErr
	}
	if !updated {
		utils.Log.Info.WithPrefix("UPDAT").Info("no new streams found")
		return nil
	}

	utils.Log.Info.WithPrefix("UPDAT").Info("found new version of toml")

	newStreamList, parseErr := parseToml(c)
	if parseErr != nil {
		return parseErr
	}
	// if new version of toml is empty, update the last update time and return
	if len(newStreamList.Streams) == 0 {
		utils.Log.Info.WithPrefix("UPDAT").Info("toml is empty")
		if setErr := c.Set(); setErr != nil {
			return setErr
		}
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
		utils.Log.Info.WithPrefix("UPDAT").Info("no new streams found")
		return nil
	}

	if insertErr := insertStreams(noDupList, c); insertErr != nil {
		return insertErr
	}
	addedCount := len(noDupList.Streams) - updateCount
	utils.Log.Info.WithPrefix("UPDAT").Infof("added %d new stream%s to database and updated %d stream%s\n", addedCount, utils.Pluralise(addedCount), updateCount, utils.Pluralise(updateCount))

	delCount, delErr := deleteStreams(newStreamList.Streams)
	if delErr != nil {
		return delErr
	}
	utils.Log.Info.WithPrefix("UPDAT").Infof("deleted %d old stream%s from database", delCount, utils.Pluralise(delCount))

	if setErr := c.Set(); setErr != nil {
		return setErr
	}
	return nil
}

// parse streams.toml and return as an s.Streams struct
func parseToml(c Config) (Streams, error) {
	utils.Log.Info.WithPrefix("UPDAT").Info("parsing toml")
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
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return 0, openErr
	}
	defer db.Close()

	sqlStmt := `
	update streams set name = ?, platform = ?, date = ?, time = ?, description = ?, url = ? where id = ?
	`

	var updateCount int
	for i, stream := range streamList.Streams {
		utils.Log.Info.WithPrefix("UPDAT").Info("updating stream", "id", stream.ID, "name", stream.Name)
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

	db, openErr := sql.Open("sqlite3", utils.Files.DB)
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
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
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
func insertStreams(streamList Streams, c Config) error {
	db, sqlErr := sql.Open("sqlite3", utils.Files.DB)
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
		utils.Log.Info.WithPrefix("UPDAT").Info("inserting stream", "name", stream.Name)
		_, insertErr := db.Exec(sqlStmt, stream.Name, stream.Platform, stream.Date, stream.Time, stream.Description, stream.URL)
		if insertErr != nil {
			return insertErr
		}
	}
	return nil
}

// delete streams from the db if the delete flag is set to true
func deleteStreams(streams []Stream) (int, error) {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return 0, openErr
	}
	defer db.Close()

	sqlStmt := `
	delete from streams where id = ?
	`
	delCount := 0
	for _, s := range streams {
		if s.Delete {
			utils.Log.Info.WithPrefix("UPDAT").Info("deleting stream", "id", s.ID, "name", s.Name)
			_, deleteErr := db.Exec(sqlStmt, s.ID)
			if deleteErr != nil {
				return 0, deleteErr
			}
			delCount++
		}
	}
	return delCount, nil
}
