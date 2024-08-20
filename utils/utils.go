/*
utils.go contains utility functions that are used throughout the bot.
*/
package utils

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"

	"gamestreams/logs"
)

// StartTime is the time the bot started.
var StartTime time.Time

// CreateTimestamp creates absolute Discord timestamps from date and time strings.
func CreateTimestamp(d string, t string) (string, string, error) {
	layout := "2006-01-02 15:04"
	if t == "" {
		dt, err := time.Parse(layout, fmt.Sprintf("%s %s", d, "09:00"))
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("<t:%d:d>", dt.Unix()), "TBC", nil
	}
	dt, err := time.Parse(layout, fmt.Sprintf("%s %s", d, t))
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("<t:%d:d>", dt.Unix()), fmt.Sprintf("<t:%d:t>", dt.Unix()), err
}

// CreateTimestampRelative returns a relative Discord timestamp from date and time
// strings. e.g. "in 2 hours"
func CreateTimestampRelative(d string, t string) (string, error) {
	layout := "2006-01-02 15:04"
	dt, err := time.Parse(layout, fmt.Sprintf("%s %s", d, t))
	return fmt.Sprintf("<t:%d:R>", dt.Unix()), err
}

// ParseTomlDate converts a date string from DD/MM/YYYY to YYYY-MM-DD.
func ParseTomlDate(d string) (string, error) {
	splitStr := strings.Split(d, "/")
	if len(splitStr) != 3 {
		return "", errors.New("invalid date format")
	}
	return fmt.Sprintf("%s-%s-%s", splitStr[2], splitStr[1], splitStr[0]), nil
}

// Pluralise returns an "s" if n is not 1. Used for pluralising words.
func Pluralise(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// RemoveSliceDuplicates removes duplicates from a slice of strings and returns a map
// of the unique strings.
func RemoveSliceDuplicates(s []string) map[string]bool {
	m := make(map[string]bool)
	for _, str := range s {
		m[str] = true
	}
	return m
}

// PlaceholderText returns "not set" if the given string is empty.
func PlaceholderText(s string) string {
	if len(s) == 0 {
		return "not set"
	}
	return s
}

// GetVideoThumbnail returns the thumbnail of a video stream from a given URL.
func GetVideoThumbnail(streamURL string) string {
	if strings.Contains(streamURL, "twitch") {
		return GetTwitchProfilePicture(streamURL)
	} else if strings.Contains(streamURL, "youtube") {
		return GetYoutubeLiveThumbnail(streamURL)
	} else if strings.Contains(streamURL, "facebook") {
		name := strings.Split(streamURL, "/")[3]
		return fmt.Sprintf("https://graph.facebook.com/%s/picture?type=large", name)
	}
	return ""
}

// GetTwitchProfilePicture returns the profile picture of a Twitch stream from a given
// URL.
func GetTwitchProfilePicture(streamURL string) string {
	doc, err := GetHTMLBody(streamURL)
	if err != nil {
		logs.LogError("UTILS", "error getting twitch html", "err", err)
		return ""
	}
	var pictureURL string
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		val, exists := s.Attr("property")
		if exists && val == "og:image" {
			pictureURL, _ = s.Attr("content")
		}
	})
	return pictureURL
}

// GetYoutubeLiveThumbnail returns the thumbnail of a YouTube stream from a given URL.
// If the URL is a direct link to the stream, it extract the video ID from the URL.
// If the URL is a channel link, it will call GetYoutubeDirectUrl to get the direct
// link and extract the video ID from that.
func GetYoutubeLiveThumbnail(streamURL string) string {
	var ID string = ""

	if strings.Contains(streamURL, "=") {
		// thumbnail from direct link
		ID = strings.Split(streamURL, "=")[1]
	} else {
		// thumbnail from channel link
		directURL, success := GetYoutubeDirectURL(streamURL)
		if success {
			ID = strings.Split(directURL, "=")[1]
		}
	}
	if ID != "" {
		return fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", ID)
	}
	return ""
}

// GetYoutubeDirectURL returns the direct URL of a YouTube stream from a profiles
// /live URL.
func GetYoutubeDirectURL(streamURL string) (string, bool) {
	var directURL string = ""
	var success bool = false

	doc, err := GetHTMLBody(streamURL)
	if err != nil {
		logs.LogError("UTILS", "error getting youtube html", "err", err)
		return "", false
	}
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		URL, _ := s.Attr("href")
		if strings.Contains(URL, "?v=") {
			directURL = URL
			success = true
		}
	})
	return directURL, success
}

// GetHTMLBody returns the HTML body of a given URL as a goquery.Document struct.
func GetHTMLBody(URL string) (*goquery.Document, error) {
	res, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, err
}

// GetUserID returns the user ID of the user who sent the interaction.
func GetUserID(i *discordgo.InteractionCreate) string {
	if i.GuildID == "" {
		return i.User.ID
	} else {
		return i.Member.User.ID
	}
}

// PatternValidator checks if a string matches a given regex pattern.
func PatternValidator(s string, pattern string) (bool, error) {
	match, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false, err
	}
	return match, nil
}

func ValidateURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}
