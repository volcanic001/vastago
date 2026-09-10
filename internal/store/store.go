package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	ErrActiveEntry = errors.New("ya hay una sesion activa")
	ErrNoActive    = errors.New("no hay una sesion activa")
)

type Entry struct {
	ID    string     `json:"id"`
	Task  string     `json:"task"`
	Note  string     `json:"note,omitempty"`
	Start time.Time  `json:"start"`
	End   *time.Time `json:"end,omitempty"`
}

func (e Entry) Duration(now time.Time) time.Duration {
	end := now
	if e.End != nil {
		end = *e.End
	}
	if end.Before(e.Start) {
		return 0
	}
	return end.Sub(e.Start)
}

func (e Entry) DurationWithin(start, end time.Time) time.Duration {
	entryEnd := end
	if e.End != nil && e.End.Before(end) {
		entryEnd = *e.End
	}
	entryStart := e.Start
	if entryStart.Before(start) {
		entryStart = start
	}
	if !entryEnd.After(entryStart) {
		return 0
	}
	return entryEnd.Sub(entryStart)
}

type Database struct {
	Entries []Entry `json:"entries"`
}

type TaskTotal struct {
	Task     string
	Duration time.Duration
	Sessions int
}

func DefaultPath() string {
	if value := strings.TrimSpace(os.Getenv("VASTAGO_DATA")); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); value != "" {
		return filepath.Join(value, "vastago", "store.json")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".", ".vastago", "store.json")
	}
	return filepath.Join(home, ".local", "share", "vastago", "store.json")
}

func Load(path string) (*Database, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Database{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("leer datos: %w", err)
	}
	var db Database
	if err := json.Unmarshal(contents, &db); err != nil {
		return nil, fmt.Errorf("decodificar datos: %w", err)
	}
	return &db, nil
}

func Save(path string, db *Database) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("crear directorio de datos: %w", err)
	}
	contents, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("codificar datos: %w", err)
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, append(contents, '\n'), 0o600); err != nil {
		return fmt.Errorf("escribir datos temporales: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("guardar datos: %w", err)
	}
	return nil
}

func (db *Database) Active() *Entry {
	for i := len(db.Entries) - 1; i >= 0; i-- {
		if db.Entries[i].End == nil {
			return &db.Entries[i]
		}
	}
	return nil
}

func (db *Database) Start(now time.Time, task, note string) (*Entry, error) {
	if db.Active() != nil {
		return nil, ErrActiveEntry
	}
	task = strings.TrimSpace(task)
	if task == "" {
		return nil, errors.New("la tarea no puede estar vacia")
	}
	entry := Entry{
		ID:    now.UTC().Format("20060102T150405.000000000"),
		Task:  task,
		Note:  strings.TrimSpace(note),
		Start: now,
	}
	db.Entries = append(db.Entries, entry)
	return &db.Entries[len(db.Entries)-1], nil
}

func (db *Database) Stop(now time.Time, note string) (*Entry, error) {
	entry := db.Active()
	if entry == nil {
		return nil, ErrNoActive
	}
	entry.End = &now
	if extra := strings.TrimSpace(note); extra != "" {
		if entry.Note != "" {
			entry.Note += " · "
		}
		entry.Note += extra
	}
	return entry, nil
}

func (db *Database) Recent(limit int) []Entry {
	if limit <= 0 || len(db.Entries) == 0 {
		return nil
	}
	start := len(db.Entries) - limit
	if start < 0 {
		start = 0
	}
	result := append([]Entry(nil), db.Entries[start:]...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func (db *Database) Since(start, now time.Time) []Entry {
	result := make([]Entry, 0)
	for _, entry := range db.Entries {
		end := now
		if entry.End != nil {
			end = *entry.End
		}
		if end.After(start) || end.Equal(start) {
			result = append(result, entry)
		}
	}
	return result
}

func Totals(entries []Entry, now time.Time) []TaskTotal {
	byTask := make(map[string]*TaskTotal)
	for _, entry := range entries {
		total, ok := byTask[entry.Task]
		if !ok {
			total = &TaskTotal{Task: entry.Task}
			byTask[entry.Task] = total
		}
		total.Duration += entry.Duration(now)
		total.Sessions++
	}
	result := make([]TaskTotal, 0, len(byTask))
	for _, total := range byTask {
		result = append(result, *total)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Duration == result[j].Duration {
			return result[i].Task < result[j].Task
		}
		return result[i].Duration > result[j].Duration
	})
	return result
}

func TotalsWithin(entries []Entry, start, end time.Time) []TaskTotal {
	byTask := make(map[string]*TaskTotal)
	for _, entry := range entries {
		duration := entry.DurationWithin(start, end)
		if duration <= 0 {
			continue
		}
		total, ok := byTask[entry.Task]
		if !ok {
			total = &TaskTotal{Task: entry.Task}
			byTask[entry.Task] = total
		}
		total.Duration += duration
		total.Sessions++
	}
	result := make([]TaskTotal, 0, len(byTask))
	for _, total := range byTask {
		result = append(result, *total)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Duration == result[j].Duration {
			return result[i].Task < result[j].Task
		}
		return result[i].Duration > result[j].Duration
	})
	return result
}

func StartOfDay(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func FormatDuration(value time.Duration) string {
	if value < 0 {
		value = 0
	}
	value = value.Round(time.Second)
	hours := int(value / time.Hour)
	minutes := int(value%time.Hour) / int(time.Minute)
	seconds := int(value%time.Minute) / int(time.Second)
	if hours > 0 {
		return fmt.Sprintf("%dh %02dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %02ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
