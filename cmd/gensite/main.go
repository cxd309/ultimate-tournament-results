// Command gensite regenerates the README table and docs/tournaments.csv
// from tournaments.csv
// tournaments.csv is hand-maintained, everything it produces is not
// run via `just index`
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gensite:", err)
		os.Exit(1)
	}
}

func run() error {
	csvPath := "tournaments.csv"
	docsCSVPath := "docs/tournaments.csv"
	readmePath := "README.md"

	tournaments, err := ReadCSV(csvPath)
	if err != nil {
		return err
	}
	if err := WriteDocsCSV(tournaments, docsCSVPath); err != nil {
		return err
	}
	if err := UpdateReadme(readmePath, tournaments); err != nil {
		return err
	}
	return nil
}

// Tournament is one row of tournaments.csv
type Tournament struct {
	Slug            string
	Event           string
	StartDate       string
	Host            string
	Version         string // -version flag tournament was archived with
	OriginalVersion string // app_version depolyment reported
	LegacyFlags     string // blank normally otherwise flags
	Notes           string // freeform notes
}

// csvHeader is written to docs/tournaments.csv
// and checked against tournaments.csv
// keep in sync with the Tournament struct field order
var csvHeader = []string{"slug", "event", "start_date", "host", "version", "original_version", "legacy_flags", "notes"}
