package logs

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/log"

	"gamestreams/config"
)

// Log is a struct that holds the loggers for the bot.
var Log Logger

// Logger is a struct that holds the loggers for the bot.
type Logger struct {
	ErrorWarn *log.Logger
	Info      *log.Logger
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
	logFile, err := os.OpenFile(config.Values.Files.Log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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
			DMOwner(dmMsg)
		} else {
			dmMsg := fmt.Sprintf("**info:**\n\tmsg=%s\n\tkeyvals:\n", msg)
			// create a string from keyvals with key=val
			for i := 0; i < len(keyvals); i += 2 {
				dmMsg += fmt.Sprintf("\t\t%s=%v\n", keyvals[i], keyvals[i+1])
			}
			DMOwner(dmMsg)
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
		DMOwner(dmMsg)
	} else {
		dmMsg := fmt.Sprintf("**error:**\n\tmsg=%s\n\tkeyvals:\n", msg)
		// create a string from keyvals with key=val
		for i := 0; i < len(keyvals); i += 2 {
			dmMsg += fmt.Sprintf("\t\t%s=%v\n", keyvals[i], keyvals[i+1])
		}
		DMOwner(dmMsg)
	}
}