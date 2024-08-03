package streams

import (
	"gamestreams/db"
	"gamestreams/logs"

)

// StreamMaintenance checks for streams in the streams table of the database that are
// over 12 months old and removes them.
func StreamMaintenance() {
	if err := db.RemoveOldStreams(); err != nil {
		logs.LogError("SCHED", "error removing old streams",
			"err", err)
	}
}
