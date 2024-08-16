package config

// Logs is a struct that holds the configuration values for the logs.
type Logs struct {
	DaysToKeep int       `toml:"days_to_keep"`
	Info       LogConfig `toml:"info"`
	Error      LogConfig `toml:"error"`
}

// LogConfig is a struct that holds the configuration values for the logs.
type LogConfig struct {
	ReportCaller    bool `toml:"report_caller"`
	CallerOffset    int  `toml:"caller_offset"`
	ReportTimestamp bool `toml:"report_timestamp"`
}
