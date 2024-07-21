package commands

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreambot/db"
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
	s.AddHandler(blacklistEdit)
	s.AddHandler(blacklistGet)
}

// listCommands is a command that lists all the owner commands
func listCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.ID != os.Getenv("OWNER_ID") {
		return
	}
	if m.Content == "!commands" || m.Content == "!help" {
		s.ChannelMessageSend(m.ChannelID, "```!uptime\n"+
			"!servercount\n"+
			"!update\n"+
			"!removeoldservers\n"+
			"!sqlx <command>\n"+
			"!streams\n"+
			"!log\n"+
			"!blacklist add <type> <id> <reason>\n"+
			"!blacklist rm <id>\n"+
			"!blacklist get```")
	}
}

// uptime is a command that returns the uptime of the bot
func uptime(s *discordgo.Session, m *discordgo.MessageCreate) {
	// check if the message author is the bot or not the owner
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != os.Getenv("OWNER_ID") ||
		len(m.Content) < 8 {
		return
	}
	if m.Content == "!uptime" {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("uptime: %s",
			time.Since(utils.StartTime).Round(time.Second).String()))
	}
}

// serverCount is a command that returns the number of servers the bot is in
func serverCount(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != os.Getenv("OWNER_ID") ||
		len(m.Content) < 12 {
		return
	}
	if m.Content == "!servercount" {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("server count: %d",
			len(s.State.Guilds)))
	}
}

// update forces an update of the streams from the streams.toml file
func update(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != os.Getenv("OWNER_ID") ||
		len(m.Content) < 8 {
		return
	}
	if m.Content == "!update" {
		var streams db.Streams
		if updateErr := streams.Update(); updateErr != nil {
			utils.LogError("OWNER", "error updating streams",
				"err", updateErr)
		} else {
			s.ChannelMessageSend(m.ChannelID, "streams updated")
		}
	}
}

// removeOldServers removes servers from the servers table that are no longer in the
// servers list
func removeOldServers(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != os.Getenv("OWNER_ID") ||
		len(m.Content) < 15 {
		return
	}
	if m.Content == "!removeoldservers" {
		if removeErr := servers.RemoveOldServerIDs(s); removeErr != nil {
			utils.LogError("OWNER", "error removing old servers",
				"err", removeErr)
		} else {
			s.ChannelMessageSend(m.ChannelID, "old servers removed")
		}
	}
}

// sqlExecute allows for the execution of SQL commands on the database via Discord
// message
func sqlExecute(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != os.Getenv("OWNER_ID") ||
		len(m.Content) < 6 {
		return
	}
	command := m.Content[:5]
	if command == "!sqlx" {
		db, openErr := sql.Open("sqlite3", utils.Files.DB)
		if openErr != nil {
			utils.LogError("OWNER", "error opening database",
				"err", openErr)
		}
		defer db.Close()
		query := m.Content[6:]
		_, execErr := db.Exec(query)
		if execErr != nil {
			utils.LogError("OWNER", "error executing database command",
				"err", execErr)
		}
		s.ChannelMessageSend(m.ChannelID, "sql executed")
	}
}

// ownerListStreams lists all upcoming streams in the streams table including their id
func ownerListStreams(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != os.Getenv("OWNER_ID") ||
		len(m.Content) < 9 {
		return
	}
	if m.Content == "!streams" {
		var streams db.Streams
		if getErr := streams.GetUpcoming(50); getErr != nil {
			utils.LogError("OWNER", "error getting streams",
				"err", getErr)
		}
		for _, stream := range streams.Streams {
			stream.ProvideUnsetValues()
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("id: `%d`\nname: `%s`\nplatform: `%s`\ndate: `%s`\ntime: `%s`\ndescription: `%s`\nurl: `%s`",
				stream.ID, stream.Name, stream.Platform, stream.Date, stream.Time, stream.Description, stream.URL))

			s.ChannelMessageSend(m.ChannelID, "----------------")
			time.Sleep(time.Second / 2)
		}
	}
}

// blacklistEdit allows the owner to add or remove users or servers from the blacklist
func blacklistEdit(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != os.Getenv("OWNER_ID") {
		return
	}
	splitString := strings.Split(m.Content, " ")
	if len(splitString) < 2 {
		s.ChannelMessageSend(m.ChannelID, "invalid command. use `!blacklist add"+
			" [type] [id] [reason]` or `!blacklist rm [id]` or `!blacklist get`")
		return
	} else if splitString[1] == "get" {
		return
	}
	command := splitString[0]
	subCommand := splitString[1]
	if command == "!blacklist" {
		if subCommand == "add" {
			blacklistAdd(s, m, splitString)
		} else if subCommand == "rm" {
			blacklistRemove(s, m, splitString)
		} else {
			s.ChannelMessageSend(m.ChannelID, "invalid command. use `!blacklist add"+
				" [type] [id] [reason]` or `!blacklist rm [id]` or `!blacklist get`")
		}
	}
}

// blacklistAdd adds a user or server to the blacklist
func blacklistAdd(s *discordgo.Session, m *discordgo.MessageCreate, splitString []string) {
	if len(splitString) < 5 || splitString[2] == "" || splitString[3] == "" ||
		splitString[4] == "" {
		s.ChannelMessageSend(m.ChannelID, "invalid command. use `!blacklist add"+
			" [type] [id] [reason]`")
		return
	}
	id_type := splitString[2]
	id := splitString[3]
	reason := strings.Join(splitString[4:], " ")
	println(reason)
	if db.AddToBlacklist(id, id_type, reason) != nil {
		utils.LogError("OWNER", "error adding to blacklist",
			"id", id,
			"id_type", id_type,
			"reason", reason)
	} else {
		s.ChannelMessageSend(m.ChannelID, "added to blacklist")
	}
}

// blacklistRemove removes a user or server from the blacklist
func blacklistRemove(s *discordgo.Session, m *discordgo.MessageCreate, splitString []string) {
	if len(splitString) < 3 || splitString[2] == "" || len(splitString) > 3 {
		s.ChannelMessageSend(m.ChannelID, "invalid command. use `!blacklist rm [id]`")
		return
	}

	id := splitString[2]
	exists, _ := db.IsBlacklisted(id, "")
	if !exists {
		s.ChannelMessageSend(m.ChannelID, "id not in blacklist")
		return
	}
	if db.RemoveFromBlacklist(id) != nil {
		utils.LogError("OWNER", "error removing from blacklist",
			"id", id)
	} else {
		s.ChannelMessageSend(m.ChannelID, "removed from blacklist")
	}
}

// blacklistGet lists all blacklisted users and servers
func blacklistGet(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != os.Getenv("OWNER_ID") ||
		len(m.Content) < 14 {
		return
	}
	if m.Content == "!blacklist get" {
		blacklist, err := db.GetBlacklist()
		if err != nil {
			utils.LogError("OWNER", "error getting blacklist",
				"err", err)
		}
		if len(blacklist) == 0 {
			s.ChannelMessageSend(m.ChannelID, "blacklist is empty")
			return
		}
		var msg string
		for _, entry := range blacklist {
			msg += fmt.Sprintf("id: `%d` id_type: `%s` date: `%s` reason: `%s`\n",
				entry.ID, entry.IDType, entry.Date, entry.Reason)
		}
		s.ChannelMessageSend(m.ChannelID, msg)
	}
}
