package utils

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

// Files is a struct that holds the file paths of important files for the bot.
var Files FilePaths

// Log is a struct that holds the loggers for the bot.
var Log Logger

// Session is a pointer to the discord session.
var Session *discordgo.Session

// StartTime is the time the bot started.
var StartTime time.Time

// FilePaths is a struct that holds the file paths of important files for the bot.
type FilePaths struct {
	DotEnv string
	DB     string
	Log    string
}

// Logger is a struct that holds the loggers for the bot.
type Logger struct {
	ErrorWarn *log.Logger
	Info      *log.Logger
}

// SetPaths sets the file paths for the bot depending on the operating system.
func (s *FilePaths) SetPaths() {
	if runtime.GOOS == "windows" {
		s.DotEnv = "config/.env"
		s.DB = "config/gamestream.db"
		s.Log = "config/gamestream.log"
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			LogError(" MAIN", "could not set filepaths", "err", err)
		}
		s.DotEnv = fmt.Sprintf("%s/config/gamestreambot/.env", home)
		s.DB = fmt.Sprintf("%s/config/gamestreambot/gamestream.db", home)
		s.Log = fmt.Sprintf("%s/config/gamestreambot/gamestream.log", home)
	}
}

// Init initializes the loggers for the bot.
func (l *Logger) Init() {
	l.Info = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
	})
	l.ErrorWarn = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
	})

	logFile, err := os.OpenFile(Files.Log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		l.ErrorWarn.WithPrefix(" MAIN").Fatal("Error opening log file",
			"err", err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	l.Info.SetOutput(mw)
	l.ErrorWarn.SetOutput(mw)
}

// LogInfo logs an info message with a prefix.
func LogInfo(prefix string, msg string, dm bool, keyvals ...interface{}) {
	Log.Info.Helper()
	Log.Info.WithPrefix(prefix).Info(msg, keyvals...)
	if dm {
		if len(keyvals) == 0 {
			dmMsg := fmt.Sprintf("**info:**\n\tmsg=%s\n",
				msg)
			DMOwner(Session, dmMsg)
		} else {
			dmMsg := fmt.Sprintf("**info:**\n\tmsg=%s\n\tkeyvals:\n", msg)
			// create a string from keyvals with key=val
			for i := 0; i < len(keyvals); i += 2 {
				dmMsg += fmt.Sprintf("\t\t%s=%v\n", keyvals[i], keyvals[i+1])
			}
			DMOwner(Session, dmMsg)
		}
	}
}

// LogError logs an error message with a prefix and sends a DM to the bot owner.
func LogError(prefix string, msg string, keyvals ...interface{}) {
	Log.ErrorWarn.Helper()
	Log.ErrorWarn.WithPrefix(prefix).Error(msg, keyvals...)

	if len(keyvals) == 0 {
		dmMsg := fmt.Sprintf("**error:**\n\tmsg=%s\n",
			msg)
		DMOwner(Session, dmMsg)
	} else {
		dmMsg := fmt.Sprintf("**error:**\n\tmsg=%s\n\tkeyvals:\n", msg)
		// create a string from keyvals with key=val
		for i := 0; i < len(keyvals); i += 2 {
			dmMsg += fmt.Sprintf("\t\t%s=%v\n", keyvals[i], keyvals[i+1])
		}
		DMOwner(Session, dmMsg)
	}
}

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
		return "", fmt.Errorf("invalid date format")
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
func GetVideoThumbnail(stream string) string {
	if strings.Contains(stream, "twitch") {
		name := strings.Split(stream, "/")[3]
		return fmt.Sprintf("https://static-cdn.jtvnw.net/previews-ttv/live_user_%s-1280x720.jpg", name)
	} else if strings.Contains(stream, "youtube") {
		return GetYoutubeLiveThumbnail(stream)
	} else if strings.Contains(stream, "facebook") {
		name := strings.Split(stream, "/")[3]
		return fmt.Sprintf("https://graph.facebook.com/%s/picture?type=large", name)
	}
	return ""
}

// GetYoutubeLiveThumbnail returns the thumbnail of a YouTube stream from a given URL.
// If the URL is a direct link to the stream, it extract the video ID from the URL.
// If the URL is a channel link, it will call GetYoutubeDirectUrl to get the direct
// link and extract the video ID from that.
func GetYoutubeLiveThumbnail(streamUrl string) string {
	var ID string = ""

	if strings.Contains(streamUrl, "=") {
		// thumbnail from direct link
		ID = strings.Split(streamUrl, "=")[1]
	} else {
		// thumbnail from channel link
		directUrl, success := GetYoutubeDirectUrl(streamUrl)
		if success {
			ID = strings.Split(directUrl, "=")[1]
		}
	}
	if ID != "" {
		return fmt.Sprintf("https://img.youtube.com/vi/%s/mqdefault.jpg", ID)
	}
	return ""
}

// GetYoutubeDirectUrl returns the direct URL of a YouTube stream from a profiles
// /live URL.
func GetYoutubeDirectUrl(streamUrl string) (string, bool) {
	var directUrl string = ""
	var success bool = false

	doc, err := GetHtmlBody(streamUrl)
	if err != nil {
		LogError(" MAIN", "error getting youtube html", "err", err)
		return "", false
	}
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Attr("href")
		if strings.Contains(url, "?v=") {
			directUrl = url
			success = true
		}
	})
	return directUrl, success
}

// GetHtmlBody returns the HTML body of a given URL as a goquery.Document struct.
func GetHtmlBody(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
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

// RegisterSession sets the global Session variable.
func RegisterSession(s *discordgo.Session) {
	Session = s
}

// TruncateLogs deletes log file entries older than 14 days by looking at
// the timestamp at the start of each line
func TruncateLogs() {
	logFile, err := os.OpenFile(Files.Log, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		LogError(" MAIN", "error opening log file", "err", err)
		return
	}
	defer logFile.Close()

	lines := []string{}
	scanner := bufio.NewScanner(logFile)
	for scanner.Scan() {
		line := scanner.Text()
		timestamp := line[0:10]
		t, err := time.Parse("2006/01/02", timestamp)
		if err != nil {
			lines = append(lines, line)
			continue
		}
		if t.After(time.Now().UTC().AddDate(0, 0, -14)) {
			lines = append(lines, line)
		}
	}

	logFile.Truncate(0)
	logFile.Seek(0, 0)
	for _, line := range lines {
		fmt.Fprintln(logFile, line)
	}
}

// IntroDM sends an introductory DM to a user when they add the bot to their server.
func IntroDM(userID string) {
	message := "🕹 Hello! Thank you for adding me to your server! 🕹\n\n" +
		"To set up your server's announcement channel, announcement role, and which platforms you want to follow, type `/settings` in the server you added me to.\n\n" +
		"For more information, type `/help`."
	LogInfo(" MAIN", "sending intro DM", false, "user", userID)

	DM(userID, message)
}

func DM(userID string, message string) {
	st, err := Session.UserChannelCreate(userID)
	if err != nil {
		LogError(" MAIN", "error creating DM channel", "err", err)
		return
	}
	_, err = Session.ChannelMessageSend(st.ID, message)
	if err != nil {
		LogError(" MAIN", "error sending DM", "err", err)
	}
}

// DM sends a direct message to the bot owner. The owner's Discord ID is stored in
// the OWNER_ID environment variable.
func DMOwner(session *discordgo.Session, message string) {
	st, err := session.UserChannelCreate(os.Getenv("OWNER_ID"))
	if err != nil {
		return
	}
	_, sendErr := session.ChannelMessageSend(st.ID, message)
	if sendErr != nil {
		return
	}
}
