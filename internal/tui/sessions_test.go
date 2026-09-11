package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/volcanic001/vastago/internal/store"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func sessionFixture(t *testing.T) Model {
	t.Helper()
	start := time.Date(2026, 9, 10, 9, 0, 0, 123, time.Local)
	end := start.Add(time.Hour)
	db := &store.Database{
		Entries: []store.Entry{{ID: "old", Task: "Lectura", Start: start, End: &end}, {ID: "new", Task: "Trabajo", Start: start, End: &end}},
		Todos:   []store.Todo{{ID: "todo", Title: "Conservar"}},
		Habits:  []store.Habit{{ID: "habit", Name: "Conservar"}},
	}
	path := filepath.Join(t.TempDir(), "store.json")
	if err := store.Save(path, db); err != nil {
		t.Fatal(err)
	}
	return Model{db: db, path: path, now: end, width: 80, height: 24, screen: sessionsScreen}
}

func pressSession(m Model, code rune, text string) Model {
	updated, _ := m.handleKey(keyMessage(code, text))
	return updated.(Model)
}

func TestSessionEditAndDeleteKeyboard(t *testing.T) {
	m := sessionFixture(t)
	m = pressSession(m, 'j', "j")
	m = pressSession(m, 'e', "e")
	if m.sessionEdit == nil || m.sessionEdit.id != "old" {
		t.Fatal("wrong selection")
	}
	m = pressSession(m, 'u', "") // no text should leave input unchanged
	m.sessionEdit.values[0] = "Lectura revisada"
	m.sessionEdit.values[1] = "Nota nueva"
	for i := 0; i < 4; i++ {
		m = pressSession(m, tea.KeyEnter, "")
	}
	if m.err != nil || m.sessionEdit != nil {
		t.Fatalf("save: %v", m.err)
	}
	db, err := store.Load(m.path)
	if err != nil {
		t.Fatal(err)
	}
	if db.Entries[0].Task != "Lectura revisada" || db.Entries[0].Note != "Nota nueva" {
		t.Fatal("edit not persisted")
	}
	if db.Entries[0].Start.Nanosecond() != 123 {
		t.Fatal("unchanged timestamp lost precision")
	}
	m = pressSession(m, 'd', "d")
	m = pressSession(m, tea.KeyEscape, "")
	if len(m.db.Entries) != 2 {
		t.Fatal("cancel deleted data")
	}
	m = pressSession(m, 'd', "d")
	m = pressSession(m, 'y', "y")
	db, err = store.Load(m.path)
	if err != nil {
		t.Fatal(err)
	}
	if len(db.Entries) != 1 || db.Entries[0].ID != "new" || len(db.Todos) != 1 || len(db.Habits) != 1 {
		t.Fatal("wrong data after deletion")
	}
}

func pressSessionCtrlS(m Model) Model {
	updated, _ := m.handleKey(tea.KeyPressMsg(tea.Key{Code: rune(115), Mod: tea.ModCtrl}))
	return updated.(Model)
}

func TestSessionEditorFieldNavigationAndTextInput(t *testing.T) {
	m := sessionFixture(t)
	m = pressSession(m, rune(101), "e")
	m = pressSession(m, tea.KeyUp, "")
	if m.sessionEdit.field != 3 {
		t.Fatalf("up field = %d, want 3", m.sessionEdit.field)
	}
	m = pressSession(m, tea.KeyDown, "")
	if m.sessionEdit.field != 0 {
		t.Fatalf("down field = %d, want 0", m.sessionEdit.field)
	}
	m = pressSession(m, tea.KeyTab, "")
	if m.sessionEdit.field != 1 {
		t.Fatalf("tab field = %d, want 1", m.sessionEdit.field)
	}
	before := m.sessionEdit.values[1]
	m = pressSession(m, rune(106), "j")
	m = pressSession(m, rune(107), "k")
	if got := m.sessionEdit.values[1]; got != before+"jk" {
		t.Fatalf("j/k did not write text: %q", got)
	}
}

func TestSessionEditorCtrlSSavesFromEveryField(t *testing.T) {
	tests := []struct {
		name  string
		field int
		edit  func(*sessionForm)
		check func(store.Entry) bool
	}{
		{"tarea", 0, func(form *sessionForm) { form.values[0] = "Trabajo editado" }, func(entry store.Entry) bool { return entry.Task == "Trabajo editado" }},
		{"nota", 1, func(form *sessionForm) { form.values[1] = "Nota editada" }, func(entry store.Entry) bool { return entry.Note == "Nota editada" }},
		{"inicio", 2, func(form *sessionForm) { form.values[2] = form.start.Add(10 * time.Minute).Format(sessionDateFormat) }, func(entry store.Entry) bool { return entry.Start.Minute() == 10 }},
		{"fin", 3, func(form *sessionForm) { form.values[3] = form.end.Add(-10 * time.Minute).Format(sessionDateFormat) }, func(entry store.Entry) bool { return entry.End != nil && entry.End.Minute() == 50 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := sessionFixture(t)
			m = pressSession(m, rune(101), "e")
			m.sessionEdit.field = test.field
			test.edit(m.sessionEdit)
			m = pressSessionCtrlS(m)
			if m.err != nil || m.sessionEdit != nil {
				t.Fatalf("ctrl+s from field %d failed: %v", test.field, m.err)
			}
			loaded, err := store.Load(m.path)
			if err != nil {
				t.Fatal(err)
			}
			if !test.check(loaded.Entries[1]) {
				t.Fatalf("edit from field %d was not persisted: %#v", test.field, loaded.Entries[1])
			}
		})
	}
}

func TestSessionEditorCtrlSKeepsInvalidForm(t *testing.T) {
	m := sessionFixture(t)
	before := append([]store.Entry(nil), m.db.Entries...)
	m = pressSession(m, rune(101), "e")
	invalidEnd := m.sessionEdit.start.Add(-time.Minute).Format(sessionDateFormat)
	m.sessionEdit.values[3] = invalidEnd
	m.sessionEdit.field = 0
	m = pressSessionCtrlS(m)
	if m.err == nil || m.sessionEdit == nil {
		t.Fatalf("invalid form was saved: err=%v", m.err)
	}
	if m.sessionEdit.values[3] != invalidEnd {
		t.Fatalf("invalid form lost changes: %q", m.sessionEdit.values[3])
	}
	if !reflect.DeepEqual(before, m.db.Entries) {
		t.Fatal("invalid save mutated sessions")
	}
}

func TestSessionEditorEscapeCancelsWithoutPersisting(t *testing.T) {
	m := sessionFixture(t)
	originalTask := m.db.Entries[1].Task
	m = pressSession(m, rune(101), "e")
	m.sessionEdit.values[0] = "No guardar"
	m = pressSession(m, tea.KeyEscape, "")
	if m.sessionEdit != nil || m.err != nil {
		t.Fatalf("escape did not close editor cleanly: %v", m.err)
	}
	loaded, err := store.Load(m.path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Entries[1].Task != originalTask {
		t.Fatalf("escape persisted task %q", loaded.Entries[1].Task)
	}
}

func TestSessionCreateAndStopPersistsValidatedTimes(t *testing.T) {
	start := time.Date(2026, 9, 11, 10, 13, 9, 123, time.Local)
	end := start.Add(47 * time.Minute)
	path := filepath.Join(t.TempDir(), "store.json")
	m := Model{db: &store.Database{}, path: path, now: start, width: 80, height: 24, screen: sessionsScreen}

	m = pressSession(m, rune(110), "n")
	m.input = []rune("Sesion nueva")
	m = pressSession(m, tea.KeyEnter, "")
	loaded, err := store.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Entries) != 1 || !loaded.Entries[0].Start.Equal(start) || loaded.Entries[0].End != nil {
		t.Fatalf("created entry = %v", loaded.Entries)
	}

	m.now = end
	m = pressSession(m, rune(120), "x")
	if m.err != nil {
		t.Fatalf("stop failed: %v", m.err)
	}
	loaded, err = store.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	entry := loaded.Entries[0]
	if entry.End == nil || !entry.Start.Equal(start) || !entry.End.Equal(end) {
		t.Fatalf("persisted start/end = %s / %v, want %s / %s", entry.Start, entry.End, start, end)
	}
	if err := store.ValidateSession(entry.Start, entry.End, end); err != nil {
		t.Fatalf("persisted session is invalid: %v", err)
	}
}

func TestRepeatSessionWithoutActivePreservesOriginal(t *testing.T) {
	m := sessionFixture(t)
	original := m.db.Entries[1]
	now := m.now.Add(15 * time.Minute)
	m.now = now
	m = pressSession(m, rune(114), "r")
	if m.err != nil {
		t.Fatalf("repeat failed: %v", m.err)
	}
	if len(m.db.Entries) != 3 {
		t.Fatalf("entries = %d, want 3", len(m.db.Entries))
	}
	if !reflect.DeepEqual(m.db.Entries[1], original) {
		t.Fatal("original session was modified")
	}
	repeated := m.db.Entries[2]
	if repeated.Task != original.Task || !repeated.Start.Equal(now) || repeated.End != nil || repeated.Note != "" {
		t.Fatalf("repeated session = %#v", repeated)
	}
	loaded, err := store.Load(m.path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Entries) != 3 || loaded.Active() == nil || loaded.Active().Task != original.Task {
		t.Fatalf("repeat was not persisted: %#v", loaded.Entries)
	}
}

func TestRepeatSessionWithActiveRequiresConfirmation(t *testing.T) {
	m := sessionFixture(t)
	activeStart := m.now.Add(5 * time.Minute)
	m.db.Entries = append(m.db.Entries, store.Entry{ID: "active", Task: "Hacer pruebas en la app", Start: activeStart})
	m.now = activeStart.Add(32*time.Minute + 14*time.Second)
	m.selectedSession = 1
	before := append([]store.Entry(nil), m.db.Entries...)

	m = pressSession(m, rune(114), "r")
	if m.sessionRepeatID != "new" {
		t.Fatalf("repeat confirmation target = %q", m.sessionRepeatID)
	}
	if !reflect.DeepEqual(m.db.Entries, before) {
		t.Fatal("requesting confirmation mutated sessions")
	}
	view := m.View().Content
	for _, want := range []string{"Ya hay una sesión activa", "Hacer pruebas en la app · 32m 14s", "Finalizarla e iniciar \"Trabajo\"", "finalizar e iniciar", "volver"} {
		if !strings.Contains(view, want) {
			t.Fatalf("confirmation missing %q: %q", want, view)
		}
	}
}

func TestConfirmRepeatStopsActiveAndStartsSelectedTask(t *testing.T) {
	m := sessionFixture(t)
	target := m.db.Entries[1]
	activeStart := m.now.Add(5 * time.Minute)
	m.db.Entries = append(m.db.Entries, store.Entry{ID: "active", Task: "Actual", Start: activeStart})
	m.now = activeStart.Add(20 * time.Minute)
	m.selectedSession = 1
	m = pressSession(m, rune(114), "r")
	m = pressSession(m, rune(121), "y")
	if m.err != nil || m.sessionRepeatID != "" {
		t.Fatalf("confirmation failed: %v", m.err)
	}
	if !reflect.DeepEqual(m.db.Entries[1], target) {
		t.Fatal("selected original session was modified")
	}
	if m.db.Entries[2].End == nil || !m.db.Entries[2].End.Equal(m.now) {
		t.Fatalf("previous active end = %v, want %v", m.db.Entries[2].End, m.now)
	}
	active := m.db.Active()
	if active == nil || active.Task != target.Task || !active.Start.Equal(m.now) || active.End != nil {
		t.Fatalf("new active session = %#v", active)
	}
	if got := activeSessionCount(m.db.Entries); got != 1 {
		t.Fatalf("active sessions = %d, want 1", got)
	}
	loaded, err := store.Load(m.path)
	if err != nil {
		t.Fatal(err)
	}
	if got := activeSessionCount(loaded.Entries); got != 1 {
		t.Fatalf("persisted active sessions = %d, want 1", got)
	}
}

func TestCancelRepeatLeavesSessionsUntouched(t *testing.T) {
	m := sessionFixture(t)
	activeStart := m.now.Add(5 * time.Minute)
	m.db.Entries = append(m.db.Entries, store.Entry{ID: "active", Task: "Actual", Start: activeStart})
	m.now = activeStart.Add(time.Minute)
	m.selectedSession = 1
	before := append([]store.Entry(nil), m.db.Entries...)
	m = pressSession(m, rune(114), "r")
	m = pressSession(m, tea.KeyEscape, "")
	if m.sessionRepeatID != "" || !reflect.DeepEqual(m.db.Entries, before) {
		t.Fatal("canceling repeat changed sessions")
	}
}

func TestRepeatSelectedActiveDoesNotDuplicateIt(t *testing.T) {
	m := sessionFixture(t)
	activeStart := m.now.Add(5 * time.Minute)
	m.db.Entries = append(m.db.Entries, store.Entry{ID: "active", Task: "Actual", Start: activeStart})
	m.now = activeStart.Add(time.Minute)
	m.selectedSession = 0
	m = pressSession(m, rune(114), "r")
	if len(m.db.Entries) != 3 || activeSessionCount(m.db.Entries) != 1 || m.sessionRepeatID != "" {
		t.Fatal("repeating selected active session created a duplicate")
	}
}

func activeSessionCount(entries []store.Entry) int {
	count := 0
	for _, entry := range entries {
		if entry.End == nil {
			count++
		}
	}
	return count
}

func TestSessionFailedSaveKeepsState(t *testing.T) {
	for _, operation := range []string{"edit", "delete"} {
		t.Run(operation, func(t *testing.T) {
			m := sessionFixture(t)
			before := append([]store.Entry(nil), m.db.Entries...)
			// An existing file used as a parent deterministically prevents saving.
			m.path = filepath.Join(m.path, "cannot-save.json")
			if operation == "edit" {
				m = pressSession(m, 'e', "e")
				m.sessionEdit.values[0] = "Must not appear"
				m.sessionEdit.field = 3
				m = pressSession(m, tea.KeyEnter, "")
			} else {
				m = pressSession(m, 'd', "d")
				m = pressSession(m, 'y', "y")
			}
			if m.err == nil {
				t.Fatal("expected save error")
			}
			if !reflect.DeepEqual(before, m.db.Entries) {
				t.Fatal("failed save mutated entries")
			}
		})
	}
}

func TestSessionViewsFitAndReachOldest(t *testing.T) {
	m := sessionFixture(t)
	for i := 0; i < 40; i++ {
		entry := m.db.Entries[0]
		entry.ID = string(rune(100 + i))
		m.db.Entries = append(m.db.Entries, entry)
	}
	for i := 0; i < 50; i++ {
		m = pressSession(m, tea.KeyDown, "")
	}
	if m.selectedSession != len(m.db.Entries)-1 {
		t.Fatal("cannot reach oldest session")
	}
	for _, size := range [][2]int{{16, 12}, {32, 16}, {60, 20}, {96, 30}} {
		for _, mode := range []string{"list", "edit", "delete"} {
			view := m
			view.width, view.height = size[0], size[1]
			if mode == "edit" {
				view = pressSession(view, 'e', "e")
			}
			if mode == "delete" {
				view = pressSession(view, 'd', "d")
			}
			content := view.View().Content
			if len(strings.Split(content, "\n")) > view.height {
				t.Fatalf("%v %s too tall", size, mode)
			}
			for _, line := range strings.Split(content, "\n") {
				if lipgloss.Width(line) > view.width {
					t.Fatalf("%v %s too wide", size, mode)
				}
			}
		}
	}
}

func TestSessionEditNormalizesDifferentOffsetsToLocalDay(t *testing.T) {
	previous := time.Local
	local := time.FixedZone("America/El_Salvador", -6*60*60)
	time.Local = local
	t.Cleanup(func() { time.Local = previous })

	start := time.Date(2026, 9, 10, 5, 29, 32, 0, time.UTC)
	originalEnd := time.Date(2026, 9, 10, 5, 30, 0, 0, time.UTC)
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, local)
	path := filepath.Join(t.TempDir(), "store.json")
	db := &store.Database{Entries: []store.Entry{{ID: "offsets", Task: "Pasada", Start: start, End: &originalEnd}}}
	if err := store.Save(path, db); err != nil {
		t.Fatal(err)
	}
	m := Model{db: db, path: path, now: now, width: 80, height: 24, screen: sessionsScreen}
	m = pressSession(m, rune(101), "e")
	if got := m.sessionEdit.values[2]; got != "2026-09-09 23:29:32" {
		t.Fatalf("local start shown in editor = %q", got)
	}
	m.sessionEdit.values[3] = "2026-09-09 23:31:00"
	m.sessionEdit.field = 3
	m = pressSession(m, tea.KeyEnter, "")
	if m.err != nil || m.sessionEdit != nil {
		t.Fatalf("same local day edit failed: %v", m.err)
	}
	loaded, err := store.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	entry := loaded.Entries[0]
	_, startOffset := entry.Start.Zone()
	_, endOffset := entry.End.Zone()
	if startOffset != 0 || endOffset != -6*60*60 {
		t.Fatalf("persisted offsets = start %d, end %d", startOffset, endOffset)
	}
	if got := entry.End.In(local).Format(sessionDateFormat); got != "2026-09-09 23:31:00" {
		t.Fatalf("persisted local end = %q", got)
	}
}

func TestLegacyInvalidSessionCanBeOpenedAndCorrected(t *testing.T) {
	start := time.Date(2026, 9, 10, 9, 0, 0, 0, time.Local)
	invalidEnd := start.Add(2 * time.Hour)
	now := start.Add(time.Hour)
	path := filepath.Join(t.TempDir(), "store.json")
	db := &store.Database{Entries: []store.Entry{{ID: "legacy", Task: "Corregir", Start: start, End: &invalidEnd}}}
	if err := store.Save(path, db); err != nil {
		t.Fatal(err)
	}
	m := Model{db: db, path: path, now: now, width: 80, height: 24, screen: sessionsScreen}
	m = pressSession(m, 'e', "e")
	if m.sessionEdit == nil {
		t.Fatal("legacy invalid session could not be opened")
	}
	for i := 0; i < 4; i++ {
		m = pressSession(m, tea.KeyEnter, "")
	}
	if m.err == nil || !strings.Contains(m.err.Error(), "futuro") {
		t.Fatalf("invalid edit error = %v", m.err)
	}
	if m.sessionEdit == nil {
		t.Fatal("invalid edit was closed")
	}

	m.sessionEdit.values[3] = start.Add(30 * time.Minute).Format(sessionDateFormat)
	m.sessionEdit.field = 3
	m = pressSession(m, tea.KeyEnter, "")
	if m.err != nil || m.sessionEdit != nil {
		t.Fatalf("corrected edit failed: err=%v edit=%v", m.err, m.sessionEdit != nil)
	}
	loaded, err := store.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Entries[0].End == nil || !loaded.Entries[0].End.Equal(start.Add(30*time.Minute)) {
		t.Fatalf("corrected session not persisted: %#v", loaded.Entries[0])
	}
}

func TestLegacyInvalidSessionCanBeDeleted(t *testing.T) {
	start := time.Date(2026, 9, 10, 9, 0, 0, 0, time.Local)
	invalidEnd := start.Add(2 * time.Hour)
	path := filepath.Join(t.TempDir(), "store.json")
	db := &store.Database{Entries: []store.Entry{{ID: "legacy", Task: "Borrar", Start: start, End: &invalidEnd}}}
	if err := store.Save(path, db); err != nil {
		t.Fatal(err)
	}
	m := Model{db: db, path: path, now: start.Add(time.Hour), width: 80, height: 24, screen: sessionsScreen}
	m = pressSession(m, 'd', "d")
	m = pressSession(m, 'y', "y")
	if m.err != nil || len(m.db.Entries) != 0 {
		t.Fatalf("legacy deletion failed: err=%v entries=%d", m.err, len(m.db.Entries))
	}
	loaded, err := store.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Entries) != 0 {
		t.Fatal("legacy session was persisted after deletion")
	}
}
