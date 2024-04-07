package db

import (
	"database/sql"

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
func (s *Streams) Query(q string) {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return
	}
	defer db.Close()

	rows, queryErr := db.Query(q)
	if queryErr != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var stream Stream
		scanErr := rows.Scan(&stream.Name, &stream.Platform, &stream.Date, &stream.Time, &stream.Description, &stream.URL)
		if scanErr != nil {
			return
		}
		s.Streams = append(s.Streams, stream)
	}
}

// GetAll gets all upcoming streams
func (s *Streams) GetUpcoming() {
	s.Query("select name, platform, date, time, description, url from streams where date >= date('now') order by date, time limit 10")
}

// GetToday gets all streams for today
func (s *Streams) GetToday() {
	s.Query("select name, platform, date, time, description, url from streams where date = date('now') and time >= time('now') order by time")
}
