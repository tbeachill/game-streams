/*
logs.go is a package that contains the loggers for the bot. The loggers are used to log
information and errors to the console and a log file. The loggers are initialized when the
bot starts and are used throughout the bot's lifecycle. The loggers are configured in the
config.toml file.
*/
package logs

import (
	"fmt"
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

// Init initializes the loggers for the bot. The options for each logger are set
// in the config.toml file.
func (l *Logger) Init() {
	l.Info = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    config.Values.Logs.Info.ReportCaller,
		CallerOffset:    config.Values.Logs.Info.CallerOffset,
		ReportTimestamp: config.Values.Logs.Info.ReportTimestamp,
	})
	l.ErrorWarn = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    config.Values.Logs.Error.ReportCaller,
		CallerOffset:    config.Values.Logs.Error.CallerOffset,
		ReportTimestamp: config.Values.Logs.Error.ReportTimestamp,
	})
	l.Info.SetOutput(os.Stdout)
	l.ErrorWarn.SetOutput(os.Stdout)
}

// LogInfo logs messages with the Info logger and a prefix. Optionally sends a DM to the
// bot owner.
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
