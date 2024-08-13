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
	// The path to the config file.
	Config string `toml:"config"`
	// The path to the database file.
	Database string `toml:"database"`
	// The path to save the database once encrypted.
	EncryptedDatabase string `toml:"encrypted_database"`
	// The path to the encryption key used to encrypt the database.
	EncryptionKey string `toml:"encryption_key"`
	// The path to the log file.
	Log string `toml:"log"`
}

// SetPaths sets the file paths for the bot depending on the operating system.
func (f *FilePaths) SetPaths() {
	file, err := os.ReadFile(f.Config)
	if err != nil {
		log.Output(5, fmt.Sprintf("CONFG: could not read config file err=%s", err))
	}
	toml.Unmarshal(file, &f)
}

// UnmarshalTOML unmarshals the TOML data into the FilePaths struct. This function
// is required as there are two different sets of paths for Windows and Linux in
// config.toml but only one struct at runtime. This prevents the TOML decoder from
// unmarshalling the paths directly into the struct.
func (f *FilePaths) UnmarshalTOML(data interface{}) error {
	m := data.(map[string]interface{})
	if w, ok := m["paths"]; ok {
		if runtime.GOOS == "windows" {
			v := w.(map[string]interface{})["windows"].(map[string]interface{})
			f.Database = v["database"].(string)
			f.EncryptedDatabase = v["encrypted_database"].(string)
			f.EncryptionKey = v["encryption_key"].(string)
			f.Log = v["log"].(string)
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				log.Output(5, fmt.Sprintf("MAIN: could not set filepaths err=%s", err))
			}
			v := w.(map[string]interface{})["linux"].(map[string]interface{})
			f.Database = fmt.Sprintf(v["database"].(string), home)
			f.EncryptedDatabase = fmt.Sprintf(v["encrypted_database"].(string), home)
			f.EncryptionKey = fmt.Sprintf(v["encryption_key"].(string), home)
			f.Log = fmt.Sprintf(v["log"].(string), home)
		}
	}
	return nil
}
