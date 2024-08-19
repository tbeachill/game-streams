/*
streams.go contains functions that are used to gather stream information from the streams
table of the database. These are middleware functions that are between the
commands and the database functions.
*/
package streams

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"gamestreams/config"
	"gamestreams/db"
	"gamestreams/logs"
	"gamestreams/utils"
)

// StreamList populates a Streams struct with upcoming streams from the streams table
// of the database. It then creates a discordgo.MessageEmbed struct with the date, time
// and title of the next [limit] streams. The limit is set in the config.toml file.
func StreamList() (*discordgo.MessageEmbed, error) {
	embed := &discordgo.MessageEmbed{
		Title: "Upcoming Streams",
		Color: config.Values.Discord.EmbedColor,
	}
	var streamList db.Streams
	if upcomErr := streamList.GetUpcoming(); upcomErr != nil {
		return nil, upcomErr
	}

	if len(streamList.Streams) == 0 {
		return embed, errors.New("no streams found")
	}
	for i, stream := range streamList.Streams {
		logs.LogInfo("STRMS", "creating embed field", false,
			"name", stream.Name,
			"time", stream.Time)

		embedField, embedErr := streamEmbedField(stream)
		if embedErr != nil {
			return nil, embedErr
		}
		//  remove the newline from the first stream embed
		if i == 0 {
			embedField.Name = strings.Split(embedField.Name, "\n")[1]
		}
		embed.Fields = append(embed.Fields, embedField)
	}
	return embed, nil
}

// StreamInfo gets a stream from the streams table of the database by name. It then
// returns a discordgo.MessageEmbed struct with the date, time, platforms, URL, and
// description of the stream.
func StreamInfo(streamName string) (*discordgo.MessageEmbed, error) {
	var streams db.Streams
	if err := streams.GetInfo(streamName); err != nil {
		return nil, err
	}
	if len(streams.Streams) == 0 {
		return nil, errors.New("no streams found")
	}
	stream := streams.Streams[0]
	date, time, dtErr := utils.CreateTimestamp(stream.Date, stream.Time)
	if dtErr != nil {
		return nil, dtErr
	}
	embed := &discordgo.MessageEmbed{
		Title: stream.Name,
		Color: config.Values.Discord.EmbedColor,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Platforms",
				Value:  stream.Platform,
				Inline: false,
			},
			{
				Name:   "\u200b\nDate",
				Value:  date,
				Inline: true,
			},
			{
				Name:   "\u200b\nTime",
				Value:  time,
				Inline: true,
			},
			{
				Name:   "\u200b\nURL",
				Value:  stream.URL,
				Inline: false,
			},
			{
				Name:   "\u200b\nDescription",
				Value:  stream.Description,
				Inline: false,
			},
		},
	}
	return embed, nil
}

// MakeStreamURLDirect checks if the stream URL is a Youtube link and if so, gets the
// direct URL to the stream. This is done as streams could be linked to as a profile's
// /live URL which would no longer link to the correct video after the stream has ended.
func MakeStreamURLDirect(stream *db.Stream) {
	if strings.Contains(stream.URL, "youtube") {
		directUrl, success := utils.GetYoutubeDirectURL(stream.URL)
		if success {
			stream.URL = directUrl
		}
	}
}

// streamEmbedField returns a discordgo.MessageEmbedField struct with the date, time,
// and name of the given stream.
func streamEmbedField(stream db.Stream) (*discordgo.MessageEmbedField, error) {
	ds, ts, tsErr := utils.CreateTimestamp(stream.Date, stream.Time)
	if tsErr != nil {
		return nil, tsErr
	}
	field := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("\u200b\n%s\t%s", ds, ts),
		Value:  stream.Name,
		Inline: false,
	}
	return field, nil
}
