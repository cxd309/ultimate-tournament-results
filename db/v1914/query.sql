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

-- name: InsertTeam :one
INSERT INTO teams (team_id, division_id, pool_id, country_id, name, abbreviation, club, seed, games_played, wins, losses, points_for, points_against, spirit_total, spirit_avg, final_standing, final_standing_calculated)
    VALUES (sqlc.arg (team_id), sqlc.arg (division_id), sqlc.arg (pool_id), sqlc.arg (country_id), sqlc.arg (name), sqlc.arg (abbreviation), sqlc.arg (club), sqlc.arg (seed), sqlc.arg (games_played), sqlc.arg (wins), sqlc.arg (losses), sqlc.arg (points_for), sqlc.arg (points_against), sqlc.arg (spirit_total), sqlc.arg (spirit_avg), sqlc.arg (final_standing), sqlc.arg (final_standing_calculated))
RETURNING
    *;

-- name: InsertPlayer :one
INSERT INTO players (player_id, team_id, first_name, last_name, jersey_num, games_played, goals, assists, callahans)
    VALUES (sqlc.arg (player_id), sqlc.arg (team_id), sqlc.arg (first_name), sqlc.arg (last_name), sqlc.arg (jersey_num), sqlc.arg (games_played), sqlc.arg (goals), sqlc.arg (assists), sqlc.arg (callahans))
RETURNING
    *;

