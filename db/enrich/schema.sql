-- Schema extension added to an existing 1.9.14-1.9.17 archive by `utr -mode enrich`,
-- which fills spirit_scores in from the ext/spiritcsv.php export of the UltiOrganizer
-- site under the Live! deployment: the one-sided scores (a game where only one team
-- has submitted) and the comments the Live! API leaves out
--
-- spirit_scores is then the whole record, read it as usual. These tables only say which
-- of its rows and comments enrich added, so livepublish/v01_09_14 leaves them out and
-- the published Live! JSON stays exactly what the API served when it was archived
--
-- one row per spirit_scores row enrich touched
-- score: the whole row was added, the Live! API didn't serve it
-- comments: the row was served without a comment, enrich added the export's
CREATE TABLE enriched_spirit (
    game_id integer NOT NULL REFERENCES games (game_id),
    team_id integer NOT NULL REFERENCES teams (team_id),
    score integer NOT NULL, -- bool
    comments integer NOT NULL, -- bool
    PRIMARY KEY (game_id, team_id),
    FOREIGN KEY (game_id, team_id) REFERENCES spirit_scores (game_id, team_id)
);

-- one row: where and when the export was read
-- its presence marks the archive as enriched, enrich runs once per archive
CREATE TABLE enrichment (
    host text NOT NULL,
    base_path text NOT NULL,
    season_id text NOT NULL,
    enriched_at text NOT NULL
);

