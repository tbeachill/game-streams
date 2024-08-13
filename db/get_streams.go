package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/robfig/cron/v3"

	"gamestreams/config"
)

// Stream is a representation of a stream. It containes information about the stream,
// the ID of the row in the streams table of the database and a flag to signify if the
// stream should be deleted.
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

// Query is a helper function to query the database using the given query string (q)
// and optional parameters. It will scan the results of the query into a Stream struct,
// appending each stream to the Streams slice of the struct.
func (s *Streams) Query(q string, params ...string) error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
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

		scanErr := rows.Scan(&stream.ID,
			&stream.Name,
			&stream.Platform,
			&stream.Date,
			&stream.Time,
			&stream.Description,
			&stream.URL)

		if scanErr != nil {
			return scanErr
		}
		s.Streams = append(s.Streams, stream)
	}
	return nil
}

// GetUpcoming gets the next 10 upcoming streams from the streams table of the database.
func (s *Streams) GetUpcoming(params ...int) error {
	var limit int
	if len(params) == 0 {
		limit = config.Values.Streams.Limit
	} else {
		limit = params[0]
	}
	if err := s.Query(`SELECT *
						FROM streams
						WHERE stream_date > DATE('now')
					UNION
						SELECT *
						FROM streams
						WHERE stream_date = DATE('now')
						AND start_time >= TIME('now')
						ORDER BY stream_date, start_time
						LIMIT ?`,
		fmt.Sprint(limit)); err != nil {
		return err
	}
	return nil
}

// GetToday gets all streams for today that have not yet started from the streams
// table of the database. It also gets streams that are scheduled for tomorrow but
// are scheduled to start before the configured stream notification cron time.
func (s *Streams) GetToday() error {
	schedule, err := cron.ParseStandard(config.Values.Schedule.StreamNotifications.Cron)
	if err != nil {
		return err
	}
	scheduleTime := schedule.Next(time.Now().UTC()).Format("15:04")
	if err := s.Query(` SELECT *
						FROM streams
						WHERE stream_date = DATE('now')
						AND start_time >= TIME('now')
					UNION
						SELECT *
						FROM streams
						WHERE stream_date = DATE('now', '+1 day')
						AND start_time < ?
						ORDER BY stream_date, start_time`,
		scheduleTime); err != nil {
		return err
	}
	return nil
}

// CheckTimeless checks for streams that are scheduled for tomorrow or the day after
// that do not have a time set. It notifies the owner which streams are missing a time
// so they can be updated.
func (s *Streams) CheckTimeless() error {
	if err := s.Query(`SELECT *
						FROM streams
						WHERE stream_date > DATE('now')
						AND stream_date <= DATE('now', '+2 day')
						AND start_time = ''`); err != nil {
		return err
	}
	return nil
}

// GetInfo gets a stream from the streams table of the database by name. It appends
// wildcard characters to the name to allow for partial matching so that the user
// does not have to type the full name of the stream.
func (s *Streams) GetInfo(name string) error {
	name = fmt.Sprintf("%%%s%%", strings.Trim(name, " "))

	if err := s.Query(`SELECT *
						FROM (
							SELECT *
							FROM streams
							WHERE stream_date = DATE('now')
							AND start_time >= TIME('now')
						UNION
							SELECT *
							FROM streams
							WHERE stream_date > DATE('now')
						)
						WHERE stream_name LIKE ? COLLATE NOCASE
						LIMIT 1`,
		name); err != nil {
		return err
	}
	return nil
}

// ProvideUnsetValues provides default values for the stream struct.
func (s *Stream) ProvideUnsetValues() {
	if s.Name == "" {
		s.Name = "NOT SET"
	}
	if s.Platform == "" {
		s.Platform = "NOT SET"
	}
	if s.Date == "" {
		s.Date = "NOT SET"
	}
	if s.Time == "" {
		s.Time = "NOT SET"
	}
	if s.Description == "" {
		s.Description = "NOT SET"
	}
	if s.URL == "" {
		s.URL = "NOT SET"
	}
}
