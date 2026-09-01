-- name: InsertTournament :exec
INSERT INTO tournament (event_name, host, season_id, base_path, app_version, start_date, end_date, timezone, status, archived_at)
    VALUES (sqlc.arg (event_name), sqlc.arg (host), sqlc.arg (season_id), sqlc.arg (base_path), sqlc.arg (app_version), sqlc.arg (start_date), sqlc.arg (end_date), sqlc.arg (timezone), sqlc.arg (status), sqlc.arg (archived_at));

-- name: GetTournament :one
SELECT
    *
FROM
    tournament;

