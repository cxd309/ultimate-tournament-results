package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

// WriteDocsCSV writes a copy of tournaments for homepage
// re-derived from the parsed rows
func WriteDocsCSV(tournaments []Tournament, path string) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close %s: %w", path, cerr)
		}
	}()

	w := csv.NewWriter(f)
	if err := w.Write(csvHeader); err != nil {
		return err
	}
	for _, t := range tournaments {
		row := []string{t.Slug, t.Event, t.StartDate, t.Host, t.Version, t.OriginalVersion, t.LegacyFlags, t.Notes}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}
