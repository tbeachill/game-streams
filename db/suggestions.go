/*
suggestions.go provides functions for interacting with the suggestions table of the
database.
*/
package db

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"gamestreams/config"
	"gamestreams/utils"
)

// Suggestion represents a row in the suggestions table of the database.
type Suggestion struct {
	// The ID of the command from the commands table.
	CommandID int
	// The name of the stream.
	Name string
	// The date the stream is scheduled for.
	Date string
	// The URL of the stream.
	URL string
}

// NewSuggestion creates a new suggestion with the given name, date, and URL. It validates
// the date and URL before creating the suggestion. If the date is not in the correct
// format or is in the past or too far in the future, it returns an error. If the URL is
// not in the correct format, it returns an error. If the date and URL are valid, it
// returns a pointer to the suggestion.
func NewSuggestion(name, date, url string) (*Suggestion, error) {
	dateCorrect, dateErr := utils.PatternValidator(date,
		`^\d{4}\-(0[1-9]|1[012])\-(0[1-9]|[12][0-9]|3[01])$`)
	if dateErr != nil || !dateCorrect {
		return nil, errors.New("date invalid or not in correct format (`YYYY-MM-DD`)")
	}
	dateTime, parseErr := time.Parse("2006-01-02", date)
	if parseErr != nil {
		return nil, errors.New("date invalid or not in correct format (`YYYY-MM-DD`)")
	}
	if dateTime.Before(time.Now()) {
		return nil, errors.New("date is in the past")
	}
	if dateTime.After(time.Now().AddDate(0, 6, 0)) {
		return nil, errors.New("date is too far in the future")
	}
	if !utils.ValidateURL(url) {
		return nil, errors.New("url is invalid")
	}
	return &Suggestion{
		Name: name,
		Date: date,
		URL:  url,
	}, nil
}

// Insert inserts the suggestion into the suggestions table of the database.
func (s *Suggestion) Insert() error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`INSERT INTO suggestions (command_id, stream_name, stream_date, stream_url)
							VALUES (?, ?, ?, ?)`,
		s.CommandID, s.Name, s.Date, s.URL)

	if execErr != nil {
		return execErr
	}
	return nil
}

// GetSuggestions gets the last [limit] suggestions from the suggestions table of the
// database. It returns a slice of Suggestion structs.
func GetSuggestions(limit int) ([]Suggestion, error) {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return nil, openErr
	}
	defer db.Close()

	rows, queryErr := db.Query(`SELECT stream_name, stream_date, stream_url
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
		scanErr := rows.Scan(&suggestion.Name, &suggestion.Date, &suggestion.URL)
		if scanErr != nil {
			return nil, scanErr
		}
		suggestions = append(suggestions, suggestion)
	}
	return suggestions, nil
}

// RemoveOldSuggestions removes suggestions that are older than the number of days
// specified in config.toml.
func RemoveOldSuggestions() error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`DELETE FROM suggestions
							WHERE command_id IN (
								SELECT id
								FROM commands
								WHERE used_date < DATETIME('now', ? || ' days')
								AND command = "suggest")`,
		-config.Values.Suggestions.DaysToKeep)

	if execErr != nil {
		return execErr
	}
	return nil
}

// ArchiveSuggestions archives suggestions that are not already in the suggestions_archive
func ArchiveSuggestions() error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`INSERT INTO suggestions_archive (stream_name, stream_date, stream_url)
							SELECT stream_name, stream_date, stream_url
							FROM suggestions
							WHERE id NOT IN (
								SELECT id
								FROM suggestions_archive
							)
						`)

	if execErr != nil {
		return execErr
	}
	return nil
}

// CountSuggestions counts the number of suggestions made by a user in the last [days] days.
func CountSuggestions(userID string, days int) (int, error) {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return 0, openErr
	}
	defer db.Close()

	row := db.QueryRow(`SELECT COUNT(*)
						FROM suggestions
						WHERE command_id IN (
							SELECT id
							FROM commands
							WHERE user_id = ?
							AND used_date > DATETIME('now', ? || ' days')
							AND command = "suggest"
						)`,
		userID, -days)

	var count int
	scanErr := row.Scan(&count)
	if scanErr != nil {
		return 0, scanErr
	}
	return count, nil
}
