package config

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/BurntSushi/toml"
)

// Values is the global configuration values for the bot.
var Values Config

// Config is a struct that holds all the configuration values for the bot.
type Config struct {
	// The bot configuration values.
	Bot Bot `toml:"bot"`
	// The file paths for the bot.
	Files FilePaths `toml:"files"`
	// Discord authentication values and configuration values.
	Discord Discord `toml:"discord"`
	// Github URLs for batch importing streams.
	Github Github `toml:"github"`
	// The Cloudflare configuration values for database backups.
	Cloudflare Cloudflare `toml:"cloudflare"`
	// The configuration values for the backup process.
	Backup Backup `toml:"backup"`
	// URLs for the various documents related to the bot.
	Documents Documents `toml:"documents"`
	// The configuration values for the streams.
	Streams Streams `toml:"streams"`
	// The configuration values for the blacklist.
	Blacklist Blacklist `toml:"blacklist"`
	// Allows cron jobs to be scheduled and enabled/disabled.
	Schedule Schedules `toml:"schedule"`
}

// LoadConfig loads the configuration values from the config.toml file into the
// Values variable.
func (c *Config) Load() {
	var configFile string
	if runtime.GOOS == "windows" {
		configFile = "config_files/config.toml"
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Output(5, fmt.Sprintf("CONFG: could not set filepaths err=%s", err))
		}
		configFile = fmt.Sprintf("%s/.config/game-streams/config.toml", home)
	}
	toml.DecodeFile(configFile, &c)
	c.Files.Config = configFile
	c.Files.SetPaths()
}
