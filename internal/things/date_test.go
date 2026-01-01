package things

import (
	"testing"
	"time"
)

func TestEncodeDecodeThingsDate(t *testing.T) {
	encoded := EncodeThingsDate(2024, 1, 15)
	year, month, day := DecodeThingsDate(encoded)
	if year != 2024 || month != 1 || day != 15 {
		t.Fatalf("unexpected decode: %d-%d-%d", year, month, day)
	}
}

func TestThingsDateToTimeAndFormat(t *testing.T) {
	loc := time.FixedZone("test", 0)
	date := EncodeThingsDate(2024, 2, 3)
	got := ThingsDateToTime(date, loc)
	if got.Format("2006-01-02") != "2024-02-03" {
		t.Fatalf("unexpected time format: %s", got.Format("2006-01-02"))
	}
	formatted := FormatThingsDate(date, loc)
	if formatted != "2024-02-03" {
		t.Fatalf("unexpected formatted date: %s", formatted)
	}
}

func TestAddDays(t *testing.T) {
	loc := time.UTC
	date := EncodeThingsDate(2024, 1, 1)
	next := AddDays(date, 2, loc)
	if next != EncodeThingsDate(2024, 1, 3) {
		t.Fatalf("unexpected date: %v", next)
	}
}
