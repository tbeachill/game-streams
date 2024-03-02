package utils

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

var ConfigFile string
var DotEnvFile string
var DBFile string
var LogFile string
var Logger *log.Logger
var EWLogger *log.Logger

type Config struct {
	CommandURL string
	StreamURL  string
	APIURL     string
	LastUpdate string
}

// create a unix timestamp from a date and time
func CreateTimestamp(d string, t string) (string, error) {
	layout := "2006-01-02 15:04"
	dt, err := time.Parse(layout, fmt.Sprintf("%s %s", d, t))

	return fmt.Sprint(dt.Unix()), err
}

func ParseTomlDate(d string) (string, error) {
	splitStr := strings.Split(d, "/")
	if len(splitStr) != 3 {
		return "", fmt.Errorf("invalid date format")
	}
	return fmt.Sprintf("%s-%s-%s", splitStr[2], splitStr[1], splitStr[0]), nil
}

func SetConfig() {
	if runtime.GOOS == "windows" {
		ConfigFile = "config/config.toml"
		DotEnvFile = "config/.env"
		DBFile = "config/gamestream.db"
		LogFile = "config/gamestream.log"
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			EWLogger.WithPrefix(" MAIN").Fatal("could not set config", "err", err)
		}
		ConfigFile = fmt.Sprintf("%s/config/gamestreambot/config.toml", home)
		DotEnvFile = fmt.Sprintf("%s/config/gamestreambot/.env", home)
		DBFile = fmt.Sprintf("%s/config/gamestreambot/gamestream.db", home)
		LogFile = fmt.Sprintf("%s/config/gamestreambot/config.toml", home)
	}
}

// create and set up the logger
func SetLogger() {
	Logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
	})
	EWLogger = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
	})

	logFile, err := os.OpenFile("config/gamestreambot.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		EWLogger.WithPrefix(" MAIN").Fatal("Error opening log file", "err", err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	Logger.SetOutput(mw)
	EWLogger.SetOutput(mw)
}
