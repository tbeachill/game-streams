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
	Bot        Bot        `toml:"bot"`
	Files      FilePaths  `toml:"files"`
	Discord    Discord    `toml:"discord"`
	Github     Github     `toml:"github"`
	Cloudflare Cloudflare `toml:"cloudflare"`
	Backup     Backup     `toml:"backup"`
	Documents  Documents  `toml:"documents"`
	Streams    Streams    `toml:"streams"`
	Blacklist  Blacklist  `toml:"blacklist"`
	Schedule   Schedules  `toml:"schedule"`
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
