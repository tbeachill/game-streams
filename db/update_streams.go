package db

import (
	"database/sql"
	"io"
	"net/http"

	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-sqlite3"

	"gamestreams/config"
	"gamestreams/logs"
	"gamestreams/utils"
)

// Update checks for new streams in the streams.toml file of the flat-files repository
// and updates the database by inserting new streams, updating existing streams, and
// deleting streams that have been marked for deletion.
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
		logs.LogInfo("UPDAT", "no new streams found", false)
		return nil
	}
	logs.LogInfo("UPDAT", "found new version of toml", false)

	*s = parseToml(c)

	// if new version of toml is empty, update the last update time and return
	if len(s.Streams) == 0 {
		logs.LogInfo("UPDAT", "toml is empty", false)
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
		logs.LogInfo("UPDAT", "no new streams found", false)
		if setErr := c.Set(); setErr != nil {
			return setErr
		}
		return nil
	}

	s.InsertStreams()

	s.DeleteStreams()

	if setErr := c.Set(); setErr != nil {
		return setErr
	}
	return nil
}

// parseToml parses the streams.toml file from the flat-files repository and returns
// as a Streams struct.
func parseToml(c Config) Streams {
	response, httpErr := http.Get(c.StreamDataURL)
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

// FormatDate runs the ParseTomlDate function on each stream in the Streams struct.
// This converts the date string from DD/MM/YYYY to YYYY-MM-DD.
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

// UpdateRow updates streams in the streams table of the database with information
// from the Streams struct. This is done when the ID of a stream in the Streams struct
// has been set to a non-zero value.
func (s *Streams) UpdateRow() error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	var updateCount int
	for i, stream := range s.Streams {
		if stream.ID != 0 {
			logs.LogInfo("UPDAT", "updating stream", false,
				"id", stream.ID,
				"name", stream.Name)

			_, updateErr := db.Exec(`UPDATE streams
									SET stream_name = ?,
										platform = ?,
										stream_date = ?,
										start_time = ?,
										stream_desc = ?,
										stream_url = ?
									WHERE id = ?`,
				stream.Name,
				stream.Platform,
				stream.Date,
				stream.Time,
				stream.Description,
				stream.URL,
				stream.ID)

			if updateErr != nil {
				return updateErr
			}
			s.Streams[i] = Stream{}
			updateCount++
		}
	}
	return nil
}

// CheckForDuplicates checks the streams table of the database for duplicates of
// streams in the Streams struct. If a a stream already exists in the streams table of
// the database, it is removed from the Streams struct.
func (s *Streams) CheckForDuplicates() error {
	rowNumber, countErr := countRows()
	if countErr != nil {
		return countErr
	}
	if rowNumber == 0 {
		return nil
	}

	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	rows, queryErr := db.Query(`SELECT stream_name,
									platform,
									stream_date,
									start_time
								FROM streams`)
	if queryErr != nil {
		return queryErr
	}
	defer rows.Close()

	var checkedList Streams
OUTER:
	for _, s := range s.Streams {
		for rows.Next() {
			var stream Stream
			scanErr := rows.Scan(&stream.Name,
				&stream.Platform,
				&stream.Date,
				&stream.Time)

			if scanErr != nil {
				return scanErr
			}
			if s.Name == stream.Name &&
				s.Platform == stream.Platform &&
				s.Date == stream.Date &&
				s.Time == stream.Time {
				continue OUTER
			}
		}
		checkedList.Streams = append(checkedList.Streams, s)
	}
	s.Streams = checkedList.Streams
	return nil
}

// countRows returns the number of rows in the streams table of the database.
func countRows() (int, error) {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return 0, openErr
	}
	defer db.Close()

	var count int

	row := db.QueryRow(`SELECT count(*)
						FROM streams`)

	scanErr := row.Scan(&count)
	if scanErr != nil {
		return 0, scanErr
	}
	return count, nil
}

// InsertStreams inserts all of the streams from the Streams struct into the streams
// table of the database.
func (s *Streams) InsertStreams() {
	db, sqlErr := sql.Open("sqlite3", config.Values.Files.Database)
	if sqlErr != nil {
		logs.LogError("UPDAT", "error opening db", "err", sqlErr)
		return
	}
	defer db.Close()

	for _, stream := range s.Streams {
		if stream.Name == "" {
			continue
		}
		logs.LogInfo("UPDAT", "inserting stream", false,
			"name", stream.Name)

		_, insertErr := db.Exec(`INSERT INTO streams
									(stream_name,
									platform,
									stream_date,
									start_time,
									stream_desc,
									stream_url)
								VALUES (?, ?, ?, ?, ?, ?)`,
			stream.Name,
			stream.Platform,
			stream.Date,
			stream.Time,
			stream.Description,
			stream.URL)

		if insertErr != nil {
			logs.LogError("UPDAT", "error inserting stream",
				"stream", stream.Name,
				"err", insertErr)

			continue
		}
	}
}

// DeleteStreams deletes streams from the streams table of the database that have been
// marked for deletion. This is done by setting the delete flag of a stream in the
// Streams struct to true.
func (s *Streams) DeleteStreams() {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		logs.LogError("UPDAT", "error opening db", "err", openErr)
		return
	}
	defer db.Close()

	for _, x := range s.Streams {
		if x.Delete {
			logs.LogInfo("UPDAT", "deleting stream", false,
				"id", x.ID,
				"name", x.Name)

			_, deleteErr := db.Exec(`DELETE FROM streams
									WHERE id = ?`,
				x.ID)

			if deleteErr != nil {
				logs.LogError("UPDAT", "error deleting stream",
					"stream", x.Name,
					"err", deleteErr)
				continue
			}
		}
	}
}

// RemoveOldStreams removes streams from the streams table of the database that are
// older than 12 months
func RemoveOldStreams() error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`DELETE FROM streams
							WHERE stream_date < date('now', '-12 months')`)
	return execErr
}
