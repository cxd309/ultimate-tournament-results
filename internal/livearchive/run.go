// Package livearchive holds the shared archive workflow for all Live! spec versions
// 1. resolve the output path
// 2. refuse to overwrite an existing archive
// 3. apply the version's schema
// 4. gather and import inside one transaction
// 5. read the result back for a sanity-checked summary
//
// No knowledge of a version's Client, Snapshot or Store types
// supplied via Deps struct
package livearchive

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
)

// ArchiveOptions carries per-run overrides for deployments that don't follow a
// version's normal conventions
// every field's zero value means "behave normally"
// currently only meaningful to v01_09_14's Client, other versions accept and
// ignore it
type ArchiveOptions struct {
	SeasonID   string // override instead of discovering it from the heartbeat
	Unprefixed bool   // static filenames carry no season-id prefix
}

// Summary is what's reported back to the caller (and printed) once Archive finishes
// first from the freshly-gathered snapshot
// then overwritten with the values read back from the committed database
// provides a sanity check that the two match
type Summary struct {
	DBPath    string
	SeasonID  string
	Name      string
	StartTime string
	Divisions int
	Pools     int
	Countries int
	Teams     int
}

// Deps is what one spec version supplies to Run for it to perform
// S is that version's Snapshot type
// Run never inspects it directly, only threads it between Gather, Summarize and Import
type Deps[S any] struct {
	// SchemaPath is the db/vXX/schema.sql file to apply to a freshly created archive
	SchemaPath string

	// Gather builds a Client for host/basePath and fetches everything the archive needs
	Gather func(ctx context.Context, host, basePath string) (S, error)

	// Summarize extracts the fields for the pre-import "gathered:" summary line
	Summarize func(snap S) Summary

	// Import writes snap to the database inside tx, via that version's Store
	Import func(ctx context.Context, tx *sql.Tx, host, basePath string, snap S) error

	// ReadBack re-reads the tournament row after commit, via that version's Store.
	ReadBack func(ctx context.Context, db *sql.DB) (name, seasonID, startTime string, err error)
}

// Run gathers one tournament from host/basePath and imports it into a fresh sqlite archive
// slug and dbPath may be empty:
// slug defaults to the gathered season id (lowercased)
// dbPath defaults to data/<slug>.db
// Archives are write-once; Run refuses to run if dbPath already exists
func Run[S any](ctx context.Context, deps Deps[S], host, basePath, slug, dbPath string) (Summary, error) {
	snap, err := deps.Gather(ctx, host, basePath)
	if err != nil {
		return Summary{}, fmt.Errorf("gather: %w", err)
	}

	summary := deps.Summarize(snap)
	fmt.Printf("gathered: season_id=%s name=%q divisions=%d pools=%d countries=%d teams=%d\n",
		summary.SeasonID, summary.Name, summary.Divisions, summary.Pools, summary.Countries, summary.Teams)

	if slug == "" {
		slug = strings.ToLower(summary.SeasonID)
	}
	if dbPath == "" {
		dbPath = fmt.Sprintf("data/%s.db", slug)
	}

	if _, err := os.Stat(dbPath); err == nil {
		return Summary{}, fmt.Errorf("%s already exists; archives are write-once, remove it first if you want to rebuild", dbPath)
	} else if !os.IsNotExist(err) {
		return Summary{}, fmt.Errorf("check %s: %w", dbPath, err)
	}

	schema, err := os.ReadFile(deps.SchemaPath)
	if err != nil {
		return Summary{}, fmt.Errorf("read schema: %w", err)
	}

	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return Summary{}, fmt.Errorf("create %s: %w", dbPath, err)
	}
	defer sqlDB.Close()

	if _, err := sqlDB.Exec(string(schema)); err != nil {
		return Summary{}, fmt.Errorf("apply schema: %w", err)
	}

	// One transaction for the whole import
	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return Summary{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if err := deps.Import(ctx, tx, host, basePath, snap); err != nil {
		return Summary{}, fmt.Errorf("import: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Summary{}, fmt.Errorf("commit: %w", err)
	}

	name, seasonID, startTime, err := deps.ReadBack(ctx, sqlDB)
	if err != nil {
		return Summary{}, fmt.Errorf("read back tournament: %w", err)
	}
	fmt.Printf("wrote %s: name=%q season_id=%s start=%s\n", dbPath, name, seasonID, startTime)

	summary.DBPath = dbPath
	summary.Name = name
	summary.SeasonID = seasonID
	summary.StartTime = startTime

	return summary, nil
}
