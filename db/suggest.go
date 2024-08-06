package db

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"gamestreams/config"
)

type Suggestion struct {
	Name      string
	Date      string
	URL       string
	DateAdded string
}

func (s *Suggestion) Insert() error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()
	s.DateAdded = time.Now().UTC().Format("2006-01-02 15:04:05")

	_, execErr := db.Exec(`INSERT INTO suggestions (stream_name, stream_date, stream_url, date_added)
							VALUES (?, ?, ?, ?)`,
		s.Name, s.Date, s.URL, s.DateAdded)
	if execErr != nil {
		return execErr
	}
	return nil
}

func GetSuggestions(limit int) ([]Suggestion, error) {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return nil, openErr
	}
	defer db.Close()

	rows, queryErr := db.Query(`SELECT stream_name, stream_date, stream_url, added_date
								FROM suggestions
								ORDER BY date_added DESC
								LIMIT ?`,
		limit)
	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()

	var suggestions []Suggestion
	for rows.Next() {
		var suggestion Suggestion
		scanErr := rows.Scan(&suggestion.Name, &suggestion.Date, &suggestion.URL, &suggestion.DateAdded)
		if scanErr != nil {
			return nil, scanErr
		}
		suggestions = append(suggestions, suggestion)
	}
	return suggestions, nil
}

// UpdateSuggestion updates the last entry in the suggestions table to include the command ID
func UpdateSuggestion() error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`UPDATE suggestions
							SET command_id = (SELECT MAX(id) FROM commands WHERE command = "suggest")`)
	if execErr != nil {
		return execErr
	}
	return nil
}

// RemoveOldSuggestions removes suggestions that are older than 30 days
func RemoveOldSuggestions() error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`DELETE FROM suggestions
							WHERE date_added < datetime('now', '-30 days')`)
	if execErr != nil {
		return execErr
	}
	return nil
}
