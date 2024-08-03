package config

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/BurntSushi/toml"
)

// FilePaths is a struct that holds the file paths of important files for the bot.
type FilePaths struct {
	Config   string `toml:"config"`
	Database string `toml:"database"`
	Log      string `toml:"log"`
}

// SetPaths sets the file paths for the bot depending on the operating system.
func (f *FilePaths) SetPaths() {
	file, err := os.ReadFile(f.Config)
	if err != nil {
		log.Output(5, fmt.Sprintf("MAIN: could not read config file err=%s", err))
	}
	toml.Unmarshal(file, &f)
}

// UnmarshalTOML unmarshals the TOML data into the FilePaths struct.
func (f *FilePaths) UnmarshalTOML(data interface{}) error {
	m := data.(map[string]interface{})
	if w, ok := m["paths"]; ok {
		if runtime.GOOS == "windows" {
			v := w.(map[string]interface{})["windows"].(map[string]interface{})
			f.Database = v["database"].(string)
			f.Log = v["log"].(string)
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				log.Output(5, fmt.Sprintf("MAIN: could not set filepaths err=%s", err))
			}
			v := w.(map[string]interface{})["linux"].(map[string]interface{})
			f.Database = fmt.Sprintf(v["database"].(string), home)
			f.Log = fmt.Sprintf(v["log"].(string), home)
		}
	}
	return nil
}
