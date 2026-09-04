package main

import (
	"fmt"
	"os"
	"strings"
)

// readmeTableHeaders is the README table's header row
// keep in sync with the cell order renderReadmeTable builds per tournament
var readmeTableHeaders = []string{"Event", "Date", "Host", "Live!", "[Legacy flags](#legacy-deployments)", "Archive"}

const (
	readmeStartMarker = "<!-- tournaments:start -->"
	readmeEndMarker   = "<!-- tournaments:end -->"
)

// UpdateReadme write (overwrites) tournaments table in readme
// inserts between tournaments markers in file
func UpdateReadme(readmePath string, tournaments []Tournament) error {
	orig, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", readmePath, err)
	}

	start := strings.Index(string(orig), readmeStartMarker)
	end := strings.Index(string(orig), readmeEndMarker)
	if start == -1 || end == -1 || end < start {
		return fmt.Errorf("%s: missing %s / %s markers", readmePath, readmeStartMarker, readmeEndMarker)
	}
	start += len(readmeStartMarker)

	var b strings.Builder
	b.WriteString(string(orig)[:start])
	b.WriteString("\n\n")
	b.WriteString(renderReadmeTable(tournaments))
	b.WriteString("\n")
	b.WriteString(string(orig)[end:])

	return os.WriteFile(readmePath, []byte(b.String()), 0o644)
}

// renderReadmeTable renders tournaments as a markdown table chronologically
// columns are left unpadded, repo is set up to use dprint formatter
func renderReadmeTable(tournaments []Tournament) string {
	var b strings.Builder
	writeRow := func(cells []string) {
		b.WriteString("| ")
		b.WriteString(strings.Join(cells, " | "))
		b.WriteString(" |\n")
	}
	writeRow(readmeTableHeaders)
	sep := make([]string, len(readmeTableHeaders))
	for i := range sep {
		sep[i] = "---"
	}
	writeRow(sep)
	for _, t := range tournaments {
		writeRow([]string{
			t.Event,
			t.StartDate,
			"`" + t.Host + "`",
			t.LiveVersion,
			legacyCell(t.LegacyFlags),
			fmt.Sprintf("[`%s`](docs/archive/%s/)", t.Slug, t.Slug),
		})
	}

	return b.String()
}

// legacyCell renders a tournament's legacy_flags for the table
// blank for a normal deployment
// each flag in its own code span otherwise, e.g. "`-season-id` `-unprefixed`"
func legacyCell(flags string) string {
	if flags == "" {
		return ""
	}
	parts := strings.Fields(flags)
	for i, p := range parts {
		parts[i] = "`" + p + "`"
	}
	return strings.Join(parts, " ")
}
