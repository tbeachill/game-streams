package db

import (
	"database/sql"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/utils"
)

type CommandData struct {
	ServerID     string
	DateTime     string
	Command      string
	Options      string
	StartTime    int64
	EndTime      int64
	ResponseTime int64
}

func (d *CommandData) Start(command discordgo.ApplicationCommandInteractionData) {
	d.StartTime = time.Now().UnixMilli()
	d.DateTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	d.Command = command.Name
	if command.Name == "streaminfo" {
		d.Options = command.Options[0].StringValue()
	}
}

func (d *CommandData) End() {
	d.EndTime = time.Now().UnixMilli()
	d.ResponseTime = d.EndTime - d.StartTime
	if err := d.DBInsert(); err != nil {
		utils.LogError("ANALYTICS", "error inserting analytics data", "err", err)
	}
}

func (d *CommandData) DBInsert() error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`INSERT INTO commands
							(server_id,
							date_time,
							command,
							options,
							response_time_ms)
						VALUES (?, ?, ?, ?, ?)`,
		d.ServerID, d.DateTime, d.Command, d.Options, d.ResponseTime)
	return execErr
}

func RemoveCommandData(serverID string) error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`DELETE FROM commands WHERE server_id = ?`, serverID)
	return execErr
}