/*
commands.go provides the functions for interactions with the database regarding commands.
*/
package db

import (
	"database/sql"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/config"
	"gamestreams/logs"
	"gamestreams/utils"
)

// CommandData is a struct that contains values for a row in the commands table of the
// database and other values used for calculating the response time of the command.
type CommandData struct {
	// The server ID where the command was used.
	ServerID string
	// The user ID of the user who used the command.
	UserID string
	// The date the command was used.
	UsedDate string
	// The time the command was used.
	UsedTime string
	// The name of the command used.
	Command string
	// The options used with the command.
	Options string
	// The time the command was started.
	StartTime int64
	// The time the command was ended.
	EndTime int64
	// The response time of the command. Calculated by subtracting the start time
	// from the end time.
	ResponseTime int64
}

// Start initializes the CommandData struct with the necessary data from the interaction.
// It sets the server ID, user ID, start time, used date, used time, command, and options.
func (d *CommandData) Start(interaction *discordgo.InteractionCreate) {
	d.ServerID = interaction.GuildID
	d.UserID = utils.GetUserID(interaction)
	d.StartTime = time.Now().UnixMilli()
	dateTime := time.Now().UTC()
	d.UsedDate = dateTime.Format("2006-01-02")
	d.UsedTime = dateTime.Format("15:04:05")
	d.Command = interaction.ApplicationCommandData().Name
	if d.Command == "streaminfo" || (d.Command == "help" &&
		len(interaction.ApplicationCommandData().Options) > 0) {
		d.Options = interaction.ApplicationCommandData().Options[0].StringValue()
	}
}

// End finalizes the CommandData struct by calculating the response time and inserting
// the data into the database.
func (d *CommandData) End() {
	d.EndTime = time.Now().UnixMilli()
	d.ResponseTime = d.EndTime - d.StartTime
	if err := d.DBInsert(); err != nil {
		logs.LogError("   DB", "error inserting command data", "err", err)
		return
	}
	// update last entry in suggestions table to include command id
	if d.Command == "suggest" {
		if err := UpdateSuggestion(); err != nil {
			logs.LogError("   DB", "error updating suggestion command_id", "err", err)
		}
	}
}

// DBInsert inserts the CommandData struct into the commands table of the database.
func (d *CommandData) DBInsert() error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`INSERT INTO commands
							(server_id,
							user_id,
							used_date,
							used_time,
							command,
							options,
							response_time_ms)
						VALUES (?, ?, ?, ?, ?, ?, ?)`,
		d.ServerID, d.UserID, d.UsedDate, d.UsedTime, d.Command, d.Options, d.ResponseTime)
	return execErr
}

// CheckUsageByUser checks the number of commands used by a user in a given period.
// Period example: "-1 day", "-1 hour", "-1 minute"
func CheckUsageByUser(userID string, period string) (int, error) {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return 0, openErr
	}
	defer db.Close()

	row := db.QueryRow(`SELECT COUNT(*)
						FROM commands
						WHERE user_id = ?
						AND used_date = DATE('now')
						AND used_time BETWEEN TIME('now', ?)
							AND TIME('now')`, userID, period)

	var count int
	scanErr := row.Scan(&count)
	return count, scanErr
}
