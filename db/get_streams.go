package db

import (
	"database/sql"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/utils"
)

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

type Streams struct {
	Streams []Stream
}

// query the database for streams by using the given query string, q, and optional parameters
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

// GetUpcoming gets all upcoming streams
func (s *Streams) GetUpcoming() error {
	if err := s.Query("select name, platform, date, time, description, url from streams where date > date('now') UNION select name, platform, date, time, description, url from streams where date = date('now') and time >= time('now') order by date, time limit 10"); err != nil {
		return err
	}
	return nil
}

// GetToday gets all streams for today
func (s *Streams) GetToday() error {
	if err := s.Query("select name, platform, date, time, description, url from streams where date = date('now') and time >= time('now') order by time"); err != nil {
		return err
	}
	return nil
}

// CheckTomorrow checks for streams tomorrow that have no time set
func (s *Streams) CheckTomorrow() error {
	if err := s.Query("select name, platform, date, time, description, url from streams where date = date('now', '+1 day') and time = ''"); err != nil {
		return err
	}
	return nil
}

// GetInfo gets information on a specific stream by name
func (s *Streams) GetInfo(name string) error {
	name = "%" + strings.Trim(name, " ") + "%"

	if err := s.Query("select name, platform, date, time, description, url from (select * from streams where date = date('now') and time >= time('now') UNION select * from streams where date > date('now')) where name like ? collate nocase limit 1", name); err != nil {
		return err
	}
	return nil
}
