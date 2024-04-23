package db

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"

	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/reports"
	"gamestreambot/utils"
)

// runner function to update the streams in the db, will return early if there are any errors
// or there are no updates
func (s *Streams) Update() error {
	var c Config

	if getErr := c.Get(); getErr != nil {
		return getErr
	}

	updated, checkErr := c.Check()
	if checkErr != nil {
		return checkErr
	}
	if !updated {
		utils.Log.Info.WithPrefix("UPDAT").Info("no new streams found")
		return nil
	}

	utils.Log.Info.WithPrefix("UPDAT").Info("found new version of toml")

	*s = parseToml(c)

	// if new version of toml is empty, update the last update time and return
	if len(s.Streams) == 0 {
		utils.Log.Info.WithPrefix("UPDAT").Info("toml is empty")
		if setErr := c.Set(); setErr != nil {
			return setErr
		}
	}

	if dateErr := s.FormatDate(); dateErr != nil {
		return dateErr
	}

	if rowErr := s.UpdateRow(); rowErr != nil {
		return rowErr
	}

	if dupErr := s.CheckForDuplicates(); dupErr != nil {
		return dupErr
	}
	if len(s.Streams) == 0 {
		utils.Log.Info.WithPrefix("UPDAT").Info("no new streams found")
		return nil
	}

	s.InsertStreams()

	s.DeleteStreams()

	if setErr := c.Set(); setErr != nil {
		return setErr
	}
	return nil
}

// parse streams.toml and return as a Streams struct
func parseToml(c Config) Streams {
	utils.Log.Info.WithPrefix("UPDAT").Info("parsing toml")
	response, httpErr := http.Get(c.StreamURL)
	if httpErr != nil {
		return Streams{}
	}
	defer response.Body.Close()

	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return Streams{}
	}
	var streamList Streams
	_, tomlErr := toml.Decode(string(body), &streamList)
	if tomlErr != nil {
		return Streams{}
	}
	return streamList
}

// format the date and time from the toml file
func (s *Streams) FormatDate() error {
	for i, stream := range s.Streams {
		d, err := utils.ParseTomlDate(stream.Date)
		if err != nil {
			return err
		}
		s.Streams[i].Date = d
	}
	return nil
}

// update an existing stream in the db
func (s *Streams) UpdateRow() error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	sqlStmt := `
	update streams set name = ?, platform = ?, date = ?, time = ?, description = ?, url = ? where id = ?
	`

	var updateCount int
	for i, stream := range s.Streams {
		utils.Log.Info.WithPrefix("UPDAT").Info("updating stream", "id", stream.ID, "name", stream.Name)
		if stream.ID != 0 {
			_, updateErr := db.Exec(sqlStmt, stream.Name, stream.Platform, stream.Date, stream.Time, stream.Description, stream.URL, stream.ID)
			if updateErr != nil {
				return updateErr
			}
			s.Streams[i] = Stream{}
			updateCount++
		}
	}
	return nil
}

// check for duplicates in the db, if a stream in the list is found in the db, go to the next row
// if none match, add the stream to a new list
func (s *Streams) CheckForDuplicates() error {
	rowNumber, countErr := countRows()
	if countErr != nil {
		return countErr
	}
	if rowNumber == 0 {
		return nil
	}

	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	sqlStmt := `
	select name, platform, date, time from streams
	`

	rows, queryErr := db.Query(sqlStmt)
	if queryErr != nil {
		return queryErr
	}
	defer rows.Close()

	var checkedList Streams
OUTER:
	for _, s := range s.Streams {
		for rows.Next() {
			var stream Stream
			scanErr := rows.Scan(&stream.Name, &stream.Platform, &stream.Date, &stream.Time)
			if scanErr != nil {
				return scanErr
			}
			if s.Name == stream.Name && s.Platform == stream.Platform && s.Date == stream.Date && s.Time == stream.Time {
				continue OUTER
			}
		}
		checkedList.Streams = append(checkedList.Streams, s)
	}
	s.Streams = checkedList.Streams
	return nil
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
func (s *Streams) InsertStreams() {
	db, sqlErr := sql.Open("sqlite3", utils.Files.DB)
	if sqlErr != nil {
		utils.Log.ErrorWarn.WithPrefix("UPDAT").Error("error opening db", "err", sqlErr)
		reports.DM(utils.Session, fmt.Sprintf("error opening db:\n\terr=%s", sqlErr))
		return
	}
	defer db.Close()

	sqlStmt := `
	insert into streams (name, platform, date, time, description, url) values (?, ?, ?, ?, ?, ?)
	`

	for _, stream := range s.Streams {
		if stream.Name == "" {
			continue
		}
		utils.Log.Info.WithPrefix("UPDAT").Info("inserting stream", "name", stream.Name)
		_, insertErr := db.Exec(sqlStmt, stream.Name, stream.Platform, stream.Date, stream.Time, stream.Description, stream.URL)
		if insertErr != nil {
			utils.Log.ErrorWarn.WithPrefix("UPDAT").Error("error inserting stream", "stream", stream.Name, "err", insertErr)
			reports.DM(utils.Session, fmt.Sprintf("error inserting stream:\n\tstream=%s\n\terr=%s", stream.Name, insertErr))
			continue
		}
	}
}

// delete streams from the db if the delete flag is set to true
func (s *Streams) DeleteStreams() {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		utils.Log.ErrorWarn.WithPrefix("UPDAT").Error("error opening db", "err", openErr)
		reports.DM(utils.Session, fmt.Sprintf("error opening db:\n\terr=%s", openErr))
		return
	}
	defer db.Close()

	sqlStmt := `
	delete from streams where id = ?
	`
	for _, x := range s.Streams {
		if x.Delete {
			utils.Log.Info.WithPrefix("UPDAT").Info("deleting stream", "id", x.ID, "name", x.Name)
			_, deleteErr := db.Exec(sqlStmt, x.ID)
			if deleteErr != nil {
				utils.Log.ErrorWarn.WithPrefix("UPDAT").Error("error deleting stream", "stream", x.Name, "err", deleteErr)
				reports.DM(utils.Session, fmt.Sprintf("error deleting stream:\n\tstream=%s\n\terr=%s", x.Name, deleteErr))
				continue
			}
		}
	}
}
