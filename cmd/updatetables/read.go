package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

// ReadCSV reads and validates tournaments.csv
func ReadCSV(path string) ([]Tournament, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("%s: empty file", path)
	}
	if got := records[0]; !equalHeader(got, csvHeader) {
		return nil, fmt.Errorf("%s: header %v, want %v", path, got, csvHeader)
	}

	tournaments := make([]Tournament, 0, len(records)-1)
	for i, row := range records[1:] {
		if len(row) != len(csvHeader) {
			return nil, fmt.Errorf("%s: row %d has %d columns, want %d", path, i+2, len(row), len(csvHeader))
		}
		if row[4] != "" && row[4] != "y" {
			return nil, fmt.Errorf("%s: row %d live is %q, want \"y\" or blank", path, i+2, row[4])
		}
		tournaments = append(tournaments, Tournament{
			Slug:            row[0],
			Event:           row[1],
			StartDate:       row[2],
			Host:            row[3],
			Live:            row[4] == "y",
			Version:         row[5],
			OriginalVersion: row[6],
			LegacyFlags:     row[7],
			Notes:           row[8],
		})
	}
	return tournaments, nil
}

func equalHeader(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
