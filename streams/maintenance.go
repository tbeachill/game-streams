package streams

import (
	"gamestreams/db"
	"gamestreams/logs"
)

// StreamMaintenance checks for streams in the streams table of the database that are
// over the limit specified in config.toml and removes them.
func StreamMaintenance() {
	if err := db.RemoveOldStreams(); err != nil {
		logs.LogError("STRMS", "error removing old streams",
			"err", err)
	}
}
