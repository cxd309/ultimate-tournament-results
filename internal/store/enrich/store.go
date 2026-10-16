// Package store is the direct interface to the tables `utr -mode enrich` adds to a
// 1.9.14-1.9.17 archive, see db/enrich/schema.sql
// spirit_scores rows themselves are inserted through store/v01_09_14
// Wraps internal/db/enrich generated queries and translates from plain Go types
package store

import (
	"context"
	"fmt"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/enrich"
)

// Store is the enrich side of the archive for one tournament, backed by a single SQLite file.
type Store struct {
	db dbgen.DBTX
	q  *dbgen.Queries
}

// New wraps an already-open database connection; the enrich tables may not exist yet
func New(db dbgen.DBTX) *Store {
	return &Store{db: db, q: dbgen.New(db)}
}

// Enriched is whether the archive has the enrich tables, i.e. enrich has already run
func (s *Store) Enriched(ctx context.Context) (bool, error) {
	var n int64
	err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'enrichment'").Scan(&n)
	if err != nil {
		return false, fmt.Errorf("check for enrichment table: %w", err)
	}
	return n > 0, nil
}

// EnrichedSpirit is the plain-Go-typed form of a single row in the enriched_spirit table
type EnrichedSpirit struct {
	GameID   int64
	TeamID   int64
	Score    bool // the whole spirit_scores row was added
	Comments bool // only its comment was added
}

func (s *Store) InsertEnrichedSpirit(ctx context.Context, e EnrichedSpirit) error {
	return s.q.InsertEnrichedSpirit(ctx, dbgen.InsertEnrichedSpiritParams{
		GameID:   e.GameID,
		TeamID:   e.TeamID,
		Score:    convert.IntBool(e.Score).Int64(),
		Comments: convert.IntBool(e.Comments).Int64(),
	})
}

// ListEnrichedSpirit returns every enriched_spirit row, none when the archive was
// never enriched
func (s *Store) ListEnrichedSpirit(ctx context.Context) ([]EnrichedSpirit, error) {
	enriched, err := s.Enriched(ctx)
	if err != nil || !enriched {
		return nil, err
	}
	rows, err := s.q.ListEnrichedSpirit(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]EnrichedSpirit, len(rows))
	for i, row := range rows {
		out[i] = EnrichedSpirit{
			GameID:   row.GameID,
			TeamID:   row.TeamID,
			Score:    row.Score != 0,
			Comments: row.Comments != 0,
		}
	}
	return out, nil
}

// SetSpiritComments sets the comments of one spirit_scores row
func (s *Store) SetSpiritComments(ctx context.Context, gameID, teamID int64, comments string) error {
	return s.q.SetSpiritComments(ctx, dbgen.SetSpiritCommentsParams{
		GameID:   gameID,
		TeamID:   teamID,
		Comments: convert.NullString(comments),
	})
}

// Enrichment is the plain-Go-typed form of the enrichment table's one row
type Enrichment struct {
	Host       string
	BasePath   string
	SeasonID   string
	EnrichedAt string
}

func (s *Store) InsertEnrichment(ctx context.Context, e Enrichment) error {
	return s.q.InsertEnrichment(ctx, dbgen.InsertEnrichmentParams{
		Host:       e.Host,
		BasePath:   e.BasePath,
		SeasonID:   e.SeasonID,
		EnrichedAt: e.EnrichedAt,
	})
}
