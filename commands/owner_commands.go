package commands

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreambot/db"
	"gamestreambot/reports"
	"gamestreambot/servers"
	"gamestreambot/utils"
)

// register event to deal with incoming messages
func RegisterOwnerCommands(s *discordgo.Session) {
	s.AddHandler(uptime)
	s.AddHandler(serverCount)
	s.AddHandler(listCommands)
	s.AddHandler(update)
	s.AddHandler(removeOldServers)
	s.AddHandler(sqlExecute)
	s.AddHandler(ownerListStreams)
}

// listCommands is a command that lists all the owner commands
func listCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.ID != os.Getenv("OWNER_ID") {
		return
	}
	if m.Content == "!commands" || m.Content == "!help" {
		s.ChannelMessageSend(m.ChannelID, "```!uptime\n!servercount\n!update\n!removeoldservers\n!sqlx <command>\n!streams\n!log```")
	}
}

// uptime is a command that returns the uptime of the bot
func uptime(s *discordgo.Session, m *discordgo.MessageCreate) {
	// check if the message author is the bot or not the owner
	if m.Author.ID == s.State.User.ID || m.Author.ID != os.Getenv("OWNER_ID") || len(m.Content) < 8 {
		return
	}
	if m.Content == "!uptime" {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("uptime: %s", time.Since(utils.StartTime).Round(time.Second).String()))
	}
}

// serverCount is a command that returns the number of servers the bot is in
func serverCount(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.ID != os.Getenv("OWNER_ID") || len(m.Content) < 12 {
		return
	}
	if m.Content == "!servercount" {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("server count: %d", len(s.State.Guilds)))
	}
}

// update forces an update of the streams from the streams.toml file
func update(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.ID != os.Getenv("OWNER_ID") || len(m.Content) < 8 {
		return
	}
	if m.Content == "!update" {
		var streams db.Streams
		if updateErr := streams.Update(); updateErr != nil {
			utils.Log.ErrorWarn.WithPrefix("UPDAT").Error("error updating streams", "err", updateErr)
			reports.DM(utils.Session, fmt.Sprintf("error updating streams:\n\terr=%s", updateErr))
		} else {
			s.ChannelMessageSend(m.ChannelID, "streams updated")
		}
	}
}

// removeOldServers removes servers from the servers table that are no longer in the servers list
func removeOldServers(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.ID != os.Getenv("OWNER_ID") || len(m.Content) < 15 {
		return
	}
	if m.Content == "!removeoldservers" {
		if removeErr := servers.RemoveOldServerIDs(s); removeErr != nil {
			utils.Log.ErrorWarn.WithPrefix("STATS").Error("error removing old servers", "err", removeErr)
			reports.DM(s, fmt.Sprintf("error removing old servers:\n\terr=%s", removeErr))
		} else {
			s.ChannelMessageSend(m.ChannelID, "old servers removed")
		}
	}
}

// sqlExecute allows for the execution of SQL commands on the database via Discord message
func sqlExecute(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.ID != os.Getenv("OWNER_ID") || len(m.Content) < 6 {
		return
	}
	if m.Content[:5] == "!sqlx" {
		db, openErr := sql.Open("sqlite3", utils.Files.DB)
		if openErr != nil {
			utils.Log.ErrorWarn.WithPrefix(" MAIN").Error("error opening db", "err", openErr)
			reports.DM(utils.Session, fmt.Sprintf("error opening db:\n\terr=%s", openErr))
		}
		defer db.Close()

		sqlStmt := m.Content[6:]
		_, execErr := db.Exec(sqlStmt)
		if execErr != nil {
			utils.Log.ErrorWarn.WithPrefix(" MAIN").Error("error executing db", "err", execErr)
			reports.DM(utils.Session, fmt.Sprintf("error executing db:\n\terr=%s", execErr))
		}
		s.ChannelMessageSend(m.ChannelID, "sql executed")
	}
}

// ownerListStreams lists all upcoming streams in the streams table including their id
func ownerListStreams(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.ID != os.Getenv("OWNER_ID") || len(m.Content) < 9 {
		return
	}
	if m.Content == "!streams" {
		var streams db.Streams
		if getErr := streams.GetUpcoming(50); getErr != nil {
			utils.Log.ErrorWarn.WithPrefix("LISTS").Error("error getting streams", "err", getErr)
			reports.DM(utils.Session, fmt.Sprintf("error getting streams:\n\terr=%s", getErr))
		}
		for _, stream := range streams.Streams {
			stream.ProvideUnsetValues()
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("id: `%d`\nname: `%s`\nplatform: `%s`\ndate: `%s`\ntime: `%s`\ndescription: `%s`\nurl: `%s`", stream.ID, stream.Name, stream.Platform, stream.Date, stream.Time, stream.Description, stream.URL))
			s.ChannelMessageSend(m.ChannelID, "----------------")
			time.Sleep(time.Second / 2)
		}
	}
}
