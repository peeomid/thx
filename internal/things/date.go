package things

import "time"

// ThingsDate encodes dates as (year<<16 | month<<12 | day<<7).
type ThingsDate int64

// EncodeThingsDate converts a date to Things' integer format.
func EncodeThingsDate(year, month, day int) ThingsDate {
	return ThingsDate((year << 16) | (month << 12) | (day << 7))
}

// DecodeThingsDate converts Things' integer format to year, month, day.
func DecodeThingsDate(date ThingsDate) (int, int, int) {
	year := int(date >> 16)
	month := int((date >> 12) & 0xF)
	day := int((date >> 7) & 0x1F)
	return year, month, day
}

// TimeToThingsDate encodes a time in local time to Things' date format.
func TimeToThingsDate(t time.Time) ThingsDate {
	year, month, day := t.Date()
	return EncodeThingsDate(year, int(month), day)
}

// ThingsDateToTime converts Things' date format into a local midnight time.
func ThingsDateToTime(date ThingsDate, loc *time.Location) time.Time {
	year, month, day := DecodeThingsDate(date)
	if loc == nil {
		loc = time.Local
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
}

// FormatThingsDate returns YYYY-MM-DD for Things date, or empty if zero.
func FormatThingsDate(date ThingsDate, loc *time.Location) string {
	if date == 0 {
		return ""
	}
	return ThingsDateToTime(date, loc).Format("2006-01-02")
}

// AddDays adds days to a Things date, returning a new Things date.
func AddDays(date ThingsDate, days int, loc *time.Location) ThingsDate {
	if date == 0 {
		return 0
	}
	return TimeToThingsDate(ThingsDateToTime(date, loc).AddDate(0, 0, days))
}
