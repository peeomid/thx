package db

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveDatabasePathDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, "Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac", "Things Database.thingsdatabase", "main.sqlite")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	resolved, err := ResolveDatabasePath("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != path {
		t.Fatalf("expected %s, got %s", path, resolved)
	}
}

func TestResolveDatabasePathLatestThingsData(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	base := filepath.Join(home, "Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac")

	first := filepath.Join(base, "ThingsData-1", "Things Database.thingsdatabase", "main.sqlite")
	second := filepath.Join(base, "ThingsData-2", "Things Database.thingsdatabase", "main.sqlite")

	if err := os.MkdirAll(filepath.Dir(first), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(second), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(first, []byte(""), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.WriteFile(second, []byte(""), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(first, past, past); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	resolved, err := ResolveDatabasePath("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != second {
		t.Fatalf("expected %s, got %s", second, resolved)
	}
}

func TestResolveDatabasePathEnv(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	custom := filepath.Join(home, "custom", "things.sqlite")
	if err := os.MkdirAll(filepath.Dir(custom), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(custom, []byte(""), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	t.Setenv(envDatabasePath, custom)

	resolved, err := ResolveDatabasePath("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != custom {
		t.Fatalf("expected %s, got %s", custom, resolved)
	}
}

func TestResolveDatabasePathPrefersThingsDataOverDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	base := filepath.Join(home, "Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac")

	defaultPath := filepath.Join(base, "Things Database.thingsdatabase", "main.sqlite")
	thingsData := filepath.Join(base, "ThingsData-1", "Things Database.thingsdatabase", "main.sqlite")

	if err := os.MkdirAll(filepath.Dir(defaultPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(thingsData), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(defaultPath, []byte(""), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.WriteFile(thingsData, []byte(""), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	resolved, err := ResolveDatabasePath("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != thingsData {
		t.Fatalf("expected %s, got %s", thingsData, resolved)
	}
}

func TestResolveDatabasePathNotFound(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	_, err := ResolveDatabasePath("", "")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrDatabaseNotFound) {
		t.Fatalf("expected ErrDatabaseNotFound")
	}
}
