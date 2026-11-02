package server

import (
	"testing"
	"time"
)

func TestFileCleanupAtFourBeijingTime(t *testing.T) {
	for _, tc := range []struct{ now, want string }{
		{"2026-11-02T03:59:59+08:00", "2026-11-02T04:00:00+08:00"},
		{"2026-11-02T04:00:00+08:00", "2026-11-03T04:00:00+08:00"},
		{"2026-12-31T23:00:00+08:00", "2027-01-01T04:00:00+08:00"},
		{"2026-11-01T20:00:00Z", "2026-11-03T04:00:00+08:00"},
	} {
		now, _ := time.Parse(time.RFC3339, tc.now)
		if got := nextFileCleanup(now).Format(time.RFC3339); got != tc.want {
			t.Errorf("%s: got %s want %s", tc.now, got, tc.want)
		}
	}
}
