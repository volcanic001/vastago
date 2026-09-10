package store

import (
	"reflect"
	"testing"
	"time"
)

func TestSessionEditValidationAndDeletion(t *testing.T) {
	start := time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	db := &Database{Entries: []Entry{{ID: "one", Task: "Before", Start: start, End: &end}, {ID: "active", Task: "Now", Start: end}}}
	before := db.Entries[0]
	early := start.Add(-time.Hour)
	if err := db.EditSession("one", "After", "note", start, &early); err == nil {
		t.Fatal("accepted negative duration")
	}
	if !reflect.DeepEqual(before, db.Entries[0]) {
		t.Fatal("invalid edit mutated entry")
	}
	if err := db.EditSession("one", "After", "note", start, nil); err == nil {
		t.Fatal("reopened closed session")
	}
	if err := db.EditSession("one", "  After  ", " note ", start, &end); err != nil {
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
