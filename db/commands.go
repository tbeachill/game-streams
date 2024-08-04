package db

import (
	"database/sql"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/config"
	"gamestreams/logs"
	"gamestreams/utils"
)

type CommandData struct {
	ServerID     string
	UserID       string
	DateTime     string
	Command      string
	Options      string
	StartTime    int64
	EndTime      int64
	ResponseTime int64
}

func (d *CommandData) Start(interaction *discordgo.InteractionCreate) {
	d.ServerID = interaction.GuildID
	d.UserID = utils.GetUserID(interaction)
	d.StartTime = time.Now().UnixMilli()
	d.DateTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	d.Command = interaction.ApplicationCommandData().Name
	if d.Command == "streaminfo" {
		d.Options = interaction.ApplicationCommandData().Options[0].StringValue()
	}
}

func (d *CommandData) End() {
	println("\n\n  {}  \n\n", d.ServerID)
	d.EndTime = time.Now().UnixMilli()
	d.ResponseTime = d.EndTime - d.StartTime
	if err := d.DBInsert(); err != nil {
		logs.LogError("ANALYTICS", "error inserting analytics data", "err", err)
	}
}

func (d *CommandData) DBInsert() error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`INSERT INTO commands
							(server_id,
							user_id,
							date_time,
							command,
							options,
							response_time_ms)
						VALUES (?, ?, ?, ?, ?, ?)`,
		d.ServerID, d.UserID, d.DateTime, d.Command, d.Options, d.ResponseTime)
	return execErr
}

func RemoveCommandData(serverID string) error {
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`DELETE FROM commands WHERE server_id = ?`, serverID)
	return execErr
}
