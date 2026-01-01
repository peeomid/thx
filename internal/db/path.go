package db

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrDatabaseNotFound = errors.New("database not found")

const envDatabasePath = "THINGSDB"

// ResolveDatabasePath finds the Things database path.
func ResolveDatabasePath(flagPath, configPath string) (string, error) {
	candidate := strings.TrimSpace(flagPath)
	if candidate == "" {
		candidate = strings.TrimSpace(configPath)
	}
	if candidate != "" {
		return resolveExistingPath(candidate)
	}

	envPath := strings.TrimSpace(os.Getenv(envDatabasePath))
	if envPath != "" {
		return resolveExistingPath(envPath)
	}

	matches, err := filepath.Glob(expandHome("~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite"))
	if err != nil {
		return "", err
	}
	if len(matches) > 0 {
		latest := matches[0]
		latestMod := time.Time{}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				continue
			}
			if info.ModTime().After(latestMod) {
				latestMod = info.ModTime()
				latest = match
			}
		}
		if latest != "" {
			return latest, nil
		}
	}

	defaultPath := expandHome("~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite")
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath, nil
	}

	return "", ErrDatabaseNotFound
}

func resolveExistingPath(path string) (string, error) {
	resolved := expandHome(path)
	if _, err := os.Stat(resolved); err == nil {
		return resolved, nil
	}
	return "", fmt.Errorf("database path not found: %s", resolved)
}

func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~/"))
}
