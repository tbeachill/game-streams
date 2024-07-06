package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/utils"
)

// Stream is a representation of a stream. It containes information about the stream, the ID of the row in the streams
// table of the database and a flag to signify if the stream should be deleted.
type Stream struct {
	ID          int
	Name        string
	Platform    string
	Date        string
	Time        string
	Description string
	URL         string
	Delete      bool
}

// Streams contains a list of Stream structs.
type Streams struct {
	Streams []Stream
}

// Query is a helper function to query the database using the given query string (q) and optional parameters.
// It will scan the results of the query into a Stream struct, appending each stream to the Streams slice of the struct.
func (s *Streams) Query(q string, params ...string) error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	var rows *sql.Rows
	if len(params) > 0 {
		rows, openErr = db.Query(q, params[0])
		if openErr != nil {
			return openErr
		}
	} else {
		rows, openErr = db.Query(q)
		if openErr != nil {
			return openErr
		}
	}
	defer rows.Close()

	for rows.Next() {
		var stream Stream
		scanErr := rows.Scan(&stream.Name, &stream.Platform, &stream.Date, &stream.Time, &stream.Description, &stream.URL)
		if scanErr != nil {
			return scanErr
		}
		s.Streams = append(s.Streams, stream)
	}
	return nil
}

// GetUpcoming gets the next 10 upcoming streams from the streams table of the database.
func (s *Streams) GetUpcoming() error {
	if err := s.Query("select name, platform, date, time, description, url from streams where date > date('now') UNION select name, platform, date, time, description, url from streams where date = date('now') and time >= time('now') order by date, time limit 10"); err != nil {
		return err
	}
	return nil
}

// GetToday gets all streams for today that have not yet started from the streams table of the database.
func (s *Streams) GetToday() error {
	if err := s.Query("select name, platform, date, time, description, url from streams where date = date('now') and time >= time('now') order by time"); err != nil {
		return err
	}
	return nil
}

// CheckTomorrow checks for streams in the streams table of the database that are scheduled for tomorrow but do not
// have a time set.
func (s *Streams) CheckTomorrow() error {
	if err := s.Query("select name, platform, date, time, description, url from streams where date = date('now', '+1 day') and time = ''"); err != nil {
		return err
	}
	return nil
}

// GetInfo gets a stream from the streams table of the database by name. It appends wildcard characters to the name to
// allow for partial matching so that the user does not have to type the full name of the stream.
func (s *Streams) GetInfo(name string) error {
	name = fmt.Sprintf("%%%s%%", strings.Trim(name, " "))

	if err := s.Query("select name, platform, date, time, description, url from (select * from streams where date = date('now') and time >= time('now') UNION select * from streams where date > date('now')) where name like ? collate nocase limit 1", name); err != nil {
		return err
	}
	return nil
}
