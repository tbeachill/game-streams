package utils

import (
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

	"gamestreambot/reports"
)

var Files FilePaths
var Log Logger
var Session *discordgo.Session

type FilePaths struct {
	DotEnv string
	DB     string
	Log    string
}

type Logger struct {
	ErrorWarn *log.Logger
	Info      *log.Logger
}

func (s *FilePaths) SetPaths() {
	if runtime.GOOS == "windows" {
		s.DotEnv = "config/.env"
		s.DB = "config/gamestream.db"
		s.Log = "config/gamestream.log"
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			Log.ErrorWarn.WithPrefix(" MAIN").Fatal("could not set filepaths", "err", err)
		}
		s.DotEnv = fmt.Sprintf("%s/config/gamestreambot/.env", home)
		s.DB = fmt.Sprintf("%s/config/gamestreambot/gamestream.db", home)
		s.Log = fmt.Sprintf("%s/config/gamestreambot/gamestream.log", home)
	}
}

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
		l.ErrorWarn.WithPrefix(" MAIN").Fatal("Error opening log file", "err", err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	l.Info.SetOutput(mw)
	l.ErrorWarn.SetOutput(mw)
}

// create a unix timestamp from a date and time
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

// create a unix timestamp from a date and time and return the relative discord time string
func CreateTimestampRelative(d string, t string) (string, error) {
	layout := "2006-01-02 15:04"
	dt, err := time.Parse(layout, fmt.Sprintf("%s %s", d, t))
	return fmt.Sprintf("<t:%d:R>", dt.Unix()), err
}

func ParseTomlDate(d string) (string, error) {
	splitStr := strings.Split(d, "/")
	if len(splitStr) != 3 {
		return "", fmt.Errorf("invalid date format")
	}
	return fmt.Sprintf("%s-%s-%s", splitStr[2], splitStr[1], splitStr[0]), nil
}

func Pluralise(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// remove duplicates from a slice of strings
func RemoveSliceDuplicates(s []string) map[string]bool {
	m := make(map[string]bool)
	for _, str := range s {
		m[str] = true
	}
	return m
}

// return a placeholder string if the input string is empty
func PlaceholderText(s string) string {
	if len(s) == 0 {
		return "not set"
	}
	return s
}

// return the url of the stream thumbnail depending on the website
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

// return the url of a youtube live stream thumbnail
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

// return the direct url of a youtube stream from its live url and a bool indicating success
func GetYoutubeDirectUrl(streamUrl string) (string, bool) {
	var directUrl string = ""
	var success bool = false

	doc, err := GetHtmlBody(streamUrl)
	if err != nil {
		Log.ErrorWarn.WithPrefix(" MAIN").Error("error getting youtube html", "err", err)
		reports.DM(Session, fmt.Sprintf("error getting youtube html:\n\terr=%s", err))
		return "", false
	}
	// will get url if
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Attr("href")
		if strings.Contains(url, "?v=") {
			directUrl = url
			success = true
		}
	})
	return directUrl, success
}

// return the html body of a webpage as a goquery doc from a url
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

// set the discord session so it is globally available
func RegisterSession(s *discordgo.Session) {
	Session = s
}

// send an intro DM to a server admin when the bot is added to a server
func IntroDM(userID string) {
	message := "🕹 Hello! Thank you for adding me to your server! 🕹\n\n" +
		"To set up your server's announcement channel, announcement role, and which platforms you want to follow, type `/settings` in the server you added me to.\n\n" +
		"For more information, type `/help`."

	Log.Info.WithPrefix(" MAIN").Info("sending intro DM", "user", userID)

	st, err := Session.UserChannelCreate(userID)
	if err != nil {
		Log.ErrorWarn.WithPrefix(" MAIN").Error("error creating intro DM channel", "err", err)
		reports.DM(Session, fmt.Sprintf("error creating intro DM channel:\n\terr=%s", err))
		return
	}
	_, err = Session.ChannelMessageSend(st.ID, message)
	if err != nil {
		Log.ErrorWarn.WithPrefix(" MAIN").Error("error sending intro DM", "err", err)
		reports.DM(Session, fmt.Sprintf("error sending intro DM:\n\terr=%s", err))
		return
	}
}
