package store

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSessionEditValidationAndDeletion(t *testing.T) {
	start := time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	now := start.Add(2 * time.Hour)
	db := &Database{Entries: []Entry{{ID: "one", Task: "Before", Start: start, End: &end}, {ID: "active", Task: "Now", Start: end}}}
	before := db.Entries[0]
	early := start.Add(-time.Hour)
	if err := db.EditSession("one", "After", "note", start, &early, now); err == nil {
		t.Fatal("accepted negative duration")
	}
	if !reflect.DeepEqual(before, db.Entries[0]) {
		t.Fatal("invalid edit mutated entry")
	}
	if err := db.EditSession("one", "After", "note", start, nil, now); err == nil {
		t.Fatal("reopened closed session")
	}
	if err := db.EditSession("one", "  After  ", " note ", start, &end, now); err != nil {
		t.Fatal(err)
	}
	if db.Entries[0].Task != "After" || db.Entries[0].Note != "note" {
		t.Fatal("edit not applied")
	}
	if err := db.DeleteSession("missing"); err != ErrSessionNotFound {
		t.Fatal(err)
	}
	if err := db.DeleteSession("one"); err != nil {
		t.Fatal(err)
	}
	if len(db.Entries) != 1 || db.Active().ID != "active" {
		t.Fatal("deleted wrong session")
	}
	if err := db.DeleteSession("active"); err != nil {
		t.Fatal(err)
	}
	if db.Active() != nil {
		t.Fatal("active session remains")
	}
}

func TestValidateSessionUsesConfiguredLocalDay(t *testing.T) {
	previous := time.Local
	local := time.FixedZone("America/El_Salvador", -6*60*60)
	time.Local = local
	t.Cleanup(func() { time.Local = previous })

	start := time.Date(2026, 9, 10, 5, 29, 32, 0, time.UTC)
	end := time.Date(2026, 9, 9, 23, 31, 0, 0, local)
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, local)
	_, startOffset := start.Zone()
	_, endOffset := end.Zone()
	if startOffset == endOffset {
		t.Fatal("test requires timestamps with different offsets")
	}
	if err := ValidateSession(start, &end, now); err != nil {
		t.Fatalf("same local day rejected: %v", err)
	}

	beforeMidnight := time.Date(2026, 9, 10, 5, 59, 0, 0, time.UTC)
	afterMidnight := time.Date(2026, 9, 10, 0, 1, 0, 0, local)
	if err := ValidateSession(beforeMidnight, &afterMidnight, now); err == nil || !strings.Contains(err.Error(), "mismo dia") {
		t.Fatalf("different local days error = %v", err)
	}
}

func TestValidateSession(t *testing.T) {
	start := time.Date(2026, 9, 10, 23, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		end  time.Time
		now  time.Time
		want string
	}{
		{"end futuro", start.Add(time.Hour), start.Add(30 * time.Minute), "futuro"},
		{"end anterior al inicio", start.Add(-time.Minute), start.Add(2 * time.Hour), "posterior"},
		{"end dia siguiente", start.Add(2 * time.Hour), start.Add(3 * time.Hour), "mismo dia"},
		{"sesion valida", start.Add(30 * time.Minute), start.Add(time.Hour), ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateSession(start, &test.end, test.now)
			if test.want == "" {
				if err != nil {
					t.Fatalf("ValidateSession() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ValidateSession() error = %v, want message containing %q", err, test.want)
			}
		})
	}
}
