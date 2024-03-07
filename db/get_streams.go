package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"gamestreambot/utils"
)

// get all future streams from the db and return as an s.Streams struct
func GetUpcomingStreams() (Streams, error) {
	utils.Logger.WithPrefix(" CMND").Info("getting upcoming streams")
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return Streams{}, openErr
	}
	defer db.Close()

	rows, queryErr := db.Query("select name, platform, date, time, description, url from streams where date >= date('now') order by date, time limit 15")
	if queryErr != nil {
		return Streams{}, queryErr
	}
	defer rows.Close()

	var streamList Streams
	for rows.Next() {
		var stream Stream
		scanErr := rows.Scan(&stream.Name, &stream.Platform, &stream.Date, &stream.Time, &stream.Description, &stream.URL)
		if scanErr != nil {
			return Streams{}, scanErr
		}
		streamList.Streams = append(streamList.Streams, stream)
	}
	utils.Logger.WithPrefix(" CMND").Info("found", "streams", len(streamList.Streams))
	return streamList, nil
}

func GetTodaysStreams() (Streams, error) {
	db, openErr := sql.Open("sqlite3", utils.DBFile)
	if openErr != nil {
		return Streams{}, openErr
	}
	defer db.Close()

	rows, queryErr := db.Query("select name, platform, date, time, description, url from streams where date = date('now') and time >= time('now') order by time")
	if queryErr != nil {
		return Streams{}, queryErr
	}
	defer rows.Close()

	var streamList Streams
	for rows.Next() {
		var stream Stream
		scanErr := rows.Scan(&stream.Name, &stream.Platform, &stream.Date, &stream.Time, &stream.Description, &stream.URL)
		if scanErr != nil {
			return Streams{}, scanErr
		}
		utils.Logger.WithPrefix("SCHED").Info("found a stream", "name", stream.Name, "time", stream.Time)
		streamList.Streams = append(streamList.Streams, stream)
	}
	return streamList, nil
}
