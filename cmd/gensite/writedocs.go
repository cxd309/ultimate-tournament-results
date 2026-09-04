package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

// WriteDocsCSV writes a copy of tournaments for homepage
// re-derived from the parsed rows
func WriteDocsCSV(tournaments []Tournament, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(csvHeader); err != nil {
		return err
	}
	for _, t := range tournaments {
		row := []string{t.Slug, t.Event, t.StartDate, t.Host, t.LiveVersion, t.LegacyFlags}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}
