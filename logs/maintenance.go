/*
maintenance.go contains functions for log file maintenance.
*/
package logs

import (
	"fmt"
	"os/exec"
	"runtime"

	"gamestreams/config"
)

// TruncateLogs truncates the logs by rotating and vacuuming journalctl logs.
// The number of days to keep logs is set in the config.toml file.
func TruncateLogs() {
	if runtime.GOOS == "linux" {
		cmd := exec.Command("sudo", "journalctl", "--rotate")
		err := cmd.Run()
		if err != nil {
			LogError(" LOGS", "error rotating journalctl logs", "err", err)
		}
		cmd = exec.Command("sudo", "journalctl", fmt.Sprintf("--vacuum-time=%dd", config.Values.Logs.DaysToKeep))
		err = cmd.Run()
		if err != nil {
			LogError(" LOGS", "error vacuuming journalctl logs", "err", err)
		}
	}
}
