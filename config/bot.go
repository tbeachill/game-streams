package config

type Bot struct {
	Version         string `toml:"version"`
	ReleaseDate     string `toml:"release_date"`
	RestoreDatabase bool   `toml:"restore_database"`
}
