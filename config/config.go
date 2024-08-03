package config

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/BurntSushi/toml"
)

var Values Config

type Config struct {
	Files      FilePaths  `toml:"files"`
	Discord    Discord    `toml:"discord"`
	Github     Github     `toml:"github"`
	Cloudflare Cloudflare `toml:"cloudflare"`
	Schedule   Schedules  `toml:"schedule"`
}

// LoadConfig loads the configuration values from the TOML file.
func (c *Config) Load() {
	var configFile string
	if runtime.GOOS == "windows" {
		configFile = "config_files/config.toml"
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Output(5, fmt.Sprintf("MAIN: could not set filepaths err=%s", err))
		}
		configFile = fmt.Sprintf("%s/.config/game-streams/config.toml", home)
	}
	toml.DecodeFile(configFile, &c)
	c.Files.Config = configFile
	c.Files.SetPaths()
}
