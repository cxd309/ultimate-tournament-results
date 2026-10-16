-- spirit_scores rows themselves are inserted through store/v01_09_14
--
-- name: SetSpiritComments :exec
UPDATE
    spirit_scores
SET
    comments = sqlc.arg (comments)
WHERE
    game_id = sqlc.arg (game_id)
    AND team_id = sqlc.arg (team_id);

-- name: InsertEnrichedSpirit :exec
INSERT INTO enriched_spirit (game_id, team_id, score, comments)
    VALUES (sqlc.arg (game_id), sqlc.arg (team_id), sqlc.arg (score), sqlc.arg (comments));

-- name: ListEnrichedSpirit :many
SELECT
    *
FROM
    enriched_spirit
ORDER BY
    game_id,
    team_id;

-- name: InsertEnrichment :exec
INSERT INTO enrichment (host, base_path, season_id, enriched_at)
    VALUES (sqlc.arg (host), sqlc.arg (base_path), sqlc.arg (season_id), sqlc.arg (enriched_at));

