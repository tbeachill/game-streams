package commands

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"gamestreams/config"
	"gamestreams/db"
	"gamestreams/logs"
	"gamestreams/servers"
	"gamestreams/utils"
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
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != config.Values.Discord.OwnerID ||
		strings.Split(m.Content, " ")[0] != "!commands" {
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
		m.Author.ID != config.Values.Discord.OwnerID ||
		strings.Split(m.Content, " ")[0] != "!uptime" {
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("uptime: %s",
		time.Since(utils.StartTime).Round(time.Second).String()))
}

// serverCount is a command that returns the number of servers the bot is in
func serverCount(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != config.Values.Discord.OwnerID ||
		strings.Split(m.Content, " ")[0] != "!servercount" {
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("server count: %d",
		len(s.State.Guilds)))
}

// update forces an update of the streams from the streams.toml file
func update(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != config.Values.Discord.OwnerID ||
		strings.Split(m.Content, " ")[0] != "!update" {
		return
	}
	var streams db.Streams
	if updateErr := streams.Update(); updateErr != nil {
		logs.LogError("OWNER", "error updating streams",
			"err", updateErr)
	} else {
		s.ChannelMessageSend(m.ChannelID, "streams updated")
	}
}

// removeOldServers removes servers from the servers table that are no longer in the
// servers list
func removeOldServers(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != config.Values.Discord.OwnerID ||
		strings.Split(m.Content, " ")[0] != "!removeoldservers" {
		return
	}
	if removeErr := servers.RemoveOldServerIDs(s); removeErr != nil {
		logs.LogError("OWNER", "error removing old servers",
			"err", removeErr)
	} else {
		s.ChannelMessageSend(m.ChannelID, "old servers removed")
	}

}

// sqlExecute allows for the execution of SQL commands on the database via Discord
// message
func sqlExecute(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != config.Values.Discord.OwnerID ||
		strings.Split(m.Content, " ")[0] != "!sqlx" {
		return
	}
	db, openErr := sql.Open("sqlite3", config.Values.Files.Database)
	if openErr != nil {
		logs.LogError("OWNER", "error opening database",
			"err", openErr)
	}
	defer db.Close()
	query := m.Content[6:]
	_, execErr := db.Exec(query)
	if execErr != nil {
		logs.LogError("OWNER", "error executing database command",
			"err", execErr)
	}
	s.ChannelMessageSend(m.ChannelID, "sql executed")
}

// ownerListStreams lists all upcoming streams in the streams table including their id
func ownerListStreams(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != config.Values.Discord.OwnerID ||
		strings.Split(m.Content, " ")[0] != "!streams" {
		return
	}
	var streams db.Streams
	if getErr := streams.GetUpcoming(50); getErr != nil {
		logs.LogError("OWNER", "error getting streams",
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

// blacklistEdit allows the owner to add or remove users or servers from the blacklist
func blacklistEdit(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != config.Values.Discord.OwnerID ||
		strings.Split(m.Content, " ")[0] != "!blacklist" {
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

	subCommand := splitString[1]
	if subCommand == "add" {
		blacklistAdd(s, m, splitString)
	} else if subCommand == "rm" || subCommand == "remove" {
		blacklistRemove(s, m, splitString)
	} else {
		s.ChannelMessageSend(m.ChannelID, "invalid command. use `!blacklist add"+
			" [type] [id] [reason]` or `!blacklist rm [id]` or `!blacklist get`")
	}
}

// blacklistAdd adds a user or server to the blacklist
func blacklistAdd(s *discordgo.Session, m *discordgo.MessageCreate, splitString []string) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != config.Values.Discord.OwnerID {
		return
	}
	if len(splitString) < 6 || splitString[2] == "" || splitString[3] == "" ||
		splitString[4] == "" || splitString[5] == "" {
		s.ChannelMessageSend(m.ChannelID, "invalid command. use `!blacklist add"+
			" [type] [id] [duration] [reason]`")
		return
	}
	idType := splitString[2]
	if idType != "user" && idType != "server" {
		s.ChannelMessageSend(m.ChannelID, "invalid type. use `user` or `server`")
		return
	}
	id := splitString[3]
	duration, convErr := strconv.Atoi(splitString[4])
	if convErr != nil {
		s.ChannelMessageSend(m.ChannelID, "invalid command. use `!blacklist add"+
			" [type] [id] [duration] [reason]`")
		return
	}
	reason := strings.Join(splitString[5:], " ")

	if dbErr := db.AddToBlacklist(id, idType, reason, duration); dbErr != nil {
		logs.LogError("OWNER", "error adding to blacklist",
			"id", id,
			"id_type", idType,
			"reason", reason,
			"duration", duration,
			"err", dbErr)
	} else {
		s.ChannelMessageSend(m.ChannelID, "added to blacklist")
	}
}

// blacklistRemove removes a user or server from the blacklist
func blacklistRemove(s *discordgo.Session, m *discordgo.MessageCreate, splitString []string) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != config.Values.Discord.OwnerID {
		return
	}
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
	if dbErr := db.RemoveFromBlacklist(id); dbErr != nil {
		logs.LogError("OWNER", "error removing from blacklist",
			"id", id,
			"err", dbErr)
	} else {
		s.ChannelMessageSend(m.ChannelID, "removed from blacklist")
	}
}

// blacklistGet lists all blacklisted users and servers
func blacklistGet(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		m.Author.ID != config.Values.Discord.OwnerID ||
		len(m.Content) < 14 {
		return
	}
	if m.Content == "!blacklist get" {
		blacklist, err := db.GetBlacklist()
		if err != nil {
			logs.LogError("OWNER", "error getting blacklist",
				"err", err)
		}
		if len(blacklist) == 0 {
			s.ChannelMessageSend(m.ChannelID, "blacklist is empty")
			return
		}
		var msg string
		for _, entry := range blacklist {
			msg += fmt.Sprintf("id: `%d` id_type: `%s` date_added: `%s` date_expires `%s` reason: `%s`\n",
				entry.ID, entry.IDType, entry.DateAdded, entry.DateExpires, entry.Reason)
		}
		s.ChannelMessageSend(m.ChannelID, msg)
	}
}
