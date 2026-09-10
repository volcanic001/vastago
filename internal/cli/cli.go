package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/volcanic001/vastago/internal/store"
	"github.com/volcanic001/vastago/internal/tui"
)

const version = "0.1.1"

func Run(args []string, stdout, stderr io.Writer) int {
	dataPath, remaining, err := globalArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	if len(remaining) == 0 {
		return runTUI(dataPath, stderr)
	}

	command, commandArgs := remaining[0], remaining[1:]
	switch command {
	case "tui":
		return runTUI(dataPath, stderr)
	case "start":
		return runStart(dataPath, commandArgs, stdout, stderr)
	case "stop":
		return runStop(dataPath, commandArgs, stdout, stderr)
	case "status":
		return runStatus(dataPath, stdout, stderr)
	case "log":
		return runLog(dataPath, commandArgs, stdout, stderr)
	case "stats":
		return runStats(dataPath, commandArgs, stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "vastago %s\n", version)
		return 0
	case "help", "--help", "-h":
		printHelp(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "error: comando desconocido %q\n\n", command)
		printHelp(stderr)
		return 2
	}
}

func globalArgs(args []string) (string, []string, error) {
	path := store.DefaultPath()
	remaining := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--data":
			if i+1 >= len(args) {
				return "", nil, errors.New("--data requiere una ruta")
			}
			path = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--data="):
			path = strings.TrimPrefix(args[i], "--data=")
		default:
			remaining = append(remaining, args[i])
		}
	}
	return path, remaining, nil
}

func runTUI(path string, stderr io.Writer) int {
	program := tea.NewProgram(tui.New(path))
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return 0
}

func runStart(path string, args []string, stdout, stderr io.Writer) int {
	note, words, err := parseNote(args)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	task := strings.TrimSpace(strings.Join(words, " "))
	if task == "" {
		fmt.Fprintln(stderr, "uso: vastago start <tarea> [--note texto]")
		return 2
	}
	db, err := store.Load(path)
	if err != nil {
		return reportError(stderr, err)
	}
	entry, err := db.Start(time.Now(), task, note)
	if err != nil {
		return reportError(stderr, err)
	}
	if err := store.Save(path, db); err != nil {
		return reportError(stderr, err)
	}
	fmt.Fprintf(stdout, "Iniciada: %s (%s)\n", entry.Task, entry.Start.Format("15:04"))
	return 0
}

func runStop(path string, args []string, stdout, stderr io.Writer) int {
	note, words, err := parseNote(args)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	if len(words) > 0 {
		fmt.Fprintln(stderr, "uso: vastago stop [--note texto]")
		return 2
	}
	db, err := store.Load(path)
	if err != nil {
		return reportError(stderr, err)
	}
	now := time.Now()
	entry, err := db.Stop(now, note)
	if err != nil {
		return reportError(stderr, err)
	}
	if err := store.Save(path, db); err != nil {
		return reportError(stderr, err)
	}
	fmt.Fprintf(stdout, "Terminada: %s · %s\n", entry.Task, store.FormatDuration(entry.Duration(now)))
	return 0
}

func parseNote(args []string) (string, []string, error) {
	var note string
	words := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		argument := args[index]
		switch {
		case argument == "--":
			words = append(words, args[index+1:]...)
			return note, words, nil
		case argument == "--note" || argument == "-n":
			if index+1 >= len(args) {
				return "", nil, fmt.Errorf("%s requiere un texto", argument)
			}
			note = args[index+1]
			index++
		case strings.HasPrefix(argument, "--note="):
			note = strings.TrimPrefix(argument, "--note=")
		case strings.HasPrefix(argument, "-"):
			return "", nil, fmt.Errorf("opcion desconocida %q", argument)
		default:
			words = append(words, argument)
		}
	}
	return note, words, nil
}

func runStatus(path string, stdout, stderr io.Writer) int {
	db, err := store.Load(path)
	if err != nil {
		return reportError(stderr, err)
	}
	entry := db.Active()
	if entry == nil {
		fmt.Fprintln(stdout, "Sin sesion activa.")
		return 0
	}
	now := time.Now()
	fmt.Fprintf(stdout, "%s · %s · desde %s\n", entry.Task, store.FormatDuration(entry.Duration(now)), entry.Start.Format("15:04"))
	return 0
}

func runLog(path string, args []string, stdout, stderr io.Writer) int {
	days, ok := parseDays("log", args, stderr)
	if !ok {
		return 2
	}
	db, err := store.Load(path)
	if err != nil {
		return reportError(stderr, err)
	}
	now := time.Now()
	start := store.StartOfDay(now).AddDate(0, 0, -(days - 1))
	entries := db.Since(start, now)
	sort.Slice(entries, func(i, j int) bool { return entries[i].Start.After(entries[j].Start) })
	if len(entries) == 0 {
		fmt.Fprintf(stdout, "No hay sesiones en los ultimos %d dias.\n", days)
		return 0
	}
	for _, entry := range entries {
		state := store.FormatDuration(entry.Duration(now))
		if entry.End == nil {
			state += " activa"
		}
		fmt.Fprintf(stdout, "%s  %-24s %s\n", entry.Start.Format("2006-01-02 15:04"), trim(entry.Task, 24), state)
		if entry.Note != "" {
			fmt.Fprintf(stdout, "                  %s\n", entry.Note)
		}
	}
	return 0
}

func runStats(path string, args []string, stdout, stderr io.Writer) int {
	days, ok := parseDays("stats", args, stderr)
	if !ok {
		return 2
	}
	db, err := store.Load(path)
	if err != nil {
		return reportError(stderr, err)
	}
	now := time.Now()
	start := store.StartOfDay(now).AddDate(0, 0, -(days - 1))
	entries := db.Since(start, now)
	totals := store.TotalsWithin(entries, start, now)
	var duration time.Duration
	for _, total := range totals {
		duration += total.Duration
	}
	fmt.Fprintf(stdout, "Ultimos %d dias · %s · %d sesiones\n", days, store.FormatDuration(duration), len(entries))
	if len(totals) == 0 {
		return 0
	}
	fmt.Fprintln(stdout)
	for _, total := range totals {
		fmt.Fprintf(stdout, "%-26s %10s  %d\n", trim(total.Task, 26), store.FormatDuration(total.Duration), total.Sessions)
	}
	return 0
}

func parseDays(name string, args []string, output io.Writer) (int, bool) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(output)
	days := flags.Int("days", 7, "numero de dias")
	flags.IntVar(days, "d", 7, "numero de dias")
	if err := flags.Parse(args); err != nil {
		return 0, false
	}
	if *days < 1 || *days > 3650 {
		fmt.Fprintln(output, "error: --days debe estar entre 1 y 3650")
		return 0, false
	}
	return *days, true
}

func reportError(output io.Writer, err error) int {
	fmt.Fprintln(output, "error:", err)
	return 1
}

func trim(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width < 2 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
}

func printHelp(output io.Writer) {
	fmt.Fprintln(output, `Vastago — tiempo enfocado, crecimiento constante

Uso:
  vastago                         abrir la interfaz TUI
  vastago start <tarea> [-n nota] iniciar una sesion
  vastago stop [-n nota]          terminar la sesion activa
  vastago status                  mostrar la sesion activa
  vastago log [-d dias]           listar sesiones recientes
  vastago stats [-d dias]         mostrar estadisticas por tarea
  vastago version                 mostrar la version

Opciones globales:
  --data <ruta>                   usar otro archivo de datos

Variables:
  VASTAGO_DATA                    ruta del archivo de datos
  XDG_DATA_HOME                   directorio base de datos`)
}
