package logs

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"gamestreams/config"
)

// TruncateLogs deletes log file entries older than the DaysToKeep value by looking at
// the timestamp at the start of each line. It also rotates and vacuums journalctl logs
// on Linux systems. The number of days to keep log entries is set in the config.toml
// file.
func TruncateLogs() {
	logFile, err := os.OpenFile(config.Values.Files.Log, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		LogError(" LOGS", "error opening log file", "err", err)
		return
	}
	defer logFile.Close()

	// rotate and vacuum journalctl logs
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

	// truncate log file
	lines := []string{}
	scanner := bufio.NewScanner(logFile)
	for scanner.Scan() {
		line := scanner.Text()
		timestamp := line[0:10]
		t, err := time.Parse("2006/01/02", timestamp)
		if err != nil {
			lines = append(lines, line)
			continue
		}
		if t.After(time.Now().UTC().AddDate(0, 0, -config.Values.Logs.DaysToKeep)) {
			lines = append(lines, line)
		}
	}

	logFile.Truncate(0)
	logFile.Seek(0, 0)
	for _, line := range lines {
		fmt.Fprintln(logFile, line)
	}
}
