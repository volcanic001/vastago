package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// configureLocalTimezone makes Go use the device timezone on Termux. Go can
// otherwise fall back to UTC there even when the shell clock is local.
func configureLocalTimezone() {
	if os.Getenv("TZ") != "" {
		return
	}
	if output, err := exec.Command("getprop", "persist.sys.timezone").Output(); err == nil {
		if name := strings.TrimSpace(string(output)); name != "" {
			if location, err := time.LoadLocation(name); err == nil {
				time.Local = location
				return
			}
		}
	}
	paths := []string{"/etc/localtime"}
	if prefix := os.Getenv("PREFIX"); prefix != "" {
		paths = append([]string{filepath.Join(prefix, "share", "zoneinfo", "localtime")}, paths...)
	}
	paths = append([]string{"/data/data/com.termux/files/usr/share/zoneinfo/localtime"}, paths...)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		location, err := time.LoadLocationFromTZData("Local", data)
		if err == nil {
			time.Local = location
			return
		}
	}
}
