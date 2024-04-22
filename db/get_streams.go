package db

import (
	"database/sql"
	"fmt"
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

// Query the db for streams
func (s *Streams) Query(q string) error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	rows, queryErr := db.Query(q)
	if queryErr != nil {
		return queryErr
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

// GetAll gets all upcoming streams
func (s *Streams) GetUpcoming() error {
	if err := s.Query("select name, platform, date, time, description, url from streams where date >= date('now') order by date, time limit 10"); err != nil {
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

// GetInfo gets information on a specific stream by name
func (s *Streams) GetInfo(name string) error {
	strings.Trim(name, " ")
	if err := s.Query(fmt.Sprintf("select name, platform, date, time, description, url from streams where name = '%s' and date >= date('now')", name)); err != nil {
		return err
	}
	return nil
}
