package analytics

import (
	"database/sql"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/utils"
)

type Data struct {
	ServerID     string
	DateTime     string
	Command      string
	Options      string
	StartTime    int64
	EndTime      int64
	ResponseTime int64
}

func (d *Data) Start(command discordgo.ApplicationCommandInteractionData) {
	d.StartTime = time.Now().UnixMilli()
	d.DateTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	d.Command = command.Name
	if command.Name == "streaminfo" {
		d.Options = command.Options[0].StringValue()
	}
}

func (d *Data) End() {
	d.EndTime = time.Now().UnixMilli()
	d.ResponseTime = d.EndTime - d.StartTime
	if err := d.DBInsert(); err != nil {
		utils.LogError("ANALYTICS", "error inserting analytics data", "err", err)
	}
}

func (d *Data) DBInsert() error {
	db, openErr := sql.Open("sqlite3", utils.Files.DB)
	if openErr != nil {
		return openErr
	}
	defer db.Close()

	_, execErr := db.Exec(`INSERT INTO analytics
							(server_id,
							date_time,
							command,
							options,
							response_time_ms)
						VALUES (?, ?, ?, ?, ?)`,
		d.ServerID, d.DateTime, d.Command, d.Options, d.ResponseTime)
	return execErr
}
