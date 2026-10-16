package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"strings"
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
		rowOrganisers := strings.Fields(row[2])
		if len(rowOrganisers) == 0 {
			return nil, fmt.Errorf("%s: row %d has no organiser", path, i+2)
		}
		for _, o := range rowOrganisers {
			if !organisers[o] {
				return nil, fmt.Errorf("%s: row %d organiser %q unknown, add it to organisers", path, i+2, o)
			}
		}
		// BULA always leads when it's one of several
		if bula := slices.Index(rowOrganisers, "BULA"); bula > 0 {
			return nil, fmt.Errorf("%s: row %d organiser %q, BULA goes first", path, i+2, row[2])
		}
		if row[5] != "" && row[5] != "y" {
			return nil, fmt.Errorf("%s: row %d live is %q, want \"y\" or blank", path, i+2, row[5])
		}
		tournaments = append(tournaments, Tournament{
			Slug:            row[0],
			Event:           row[1],
			Organisers:      rowOrganisers,
			StartDate:       row[3],
			Host:            row[4],
			Live:            row[5] == "y",
			Version:         row[6],
			OriginalVersion: row[7],
			LegacyFlags:     row[8],
			Notes:           row[9],
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
