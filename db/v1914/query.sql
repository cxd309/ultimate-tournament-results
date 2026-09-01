-- name: InsertTournament :exec
INSERT INTO tournament (event_name, host, season_id, base_path, app_version, start_date, end_date, timezone, status, archived_at)
    VALUES (sqlc.arg (event_name), sqlc.arg (host), sqlc.arg (season_id), sqlc.arg (base_path), sqlc.arg (app_version), sqlc.arg (start_date), sqlc.arg (end_date), sqlc.arg (timezone), sqlc.arg (status), sqlc.arg (archived_at));

-- name: GetTournament :one
SELECT
    *
FROM
    tournament;

-- name: InsertDivision :one
INSERT INTO divisions (series_id, name, ordering)
    VALUES (sqlc.arg (series_id), sqlc.arg (name), sqlc.arg (ordering))
RETURNING
    *;

-- name: InsertPool :one
INSERT INTO pools (pool_id, division_id, name, ordering, pool_type)
    VALUES (sqlc.arg (pool_id), sqlc.arg (division_id), sqlc.arg (name), sqlc.arg (ordering), sqlc.arg (pool_type))
RETURNING
    *;

-- name: InsertCountry :one
INSERT INTO countries (country_ext_id, name, abbreviation, flag_file)
    VALUES (sqlc.arg (country_ext_id), sqlc.arg (name), sqlc.arg (abbreviation), sqlc.arg (flag_file))
RETURNING
    *;

