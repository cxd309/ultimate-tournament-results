-- name: InsertTournament :exec
INSERT INTO tournament (season_id, name, starttime, endtime, timezone, host, base_path, app_version, archived_at)
    VALUES (sqlc.arg (season_id), sqlc.arg (name), sqlc.arg (starttime), sqlc.arg (endtime), sqlc.arg (timezone), sqlc.arg (host), sqlc.arg (base_path), sqlc.arg (app_version), sqlc.arg (archived_at));

-- name: GetTournament :one
SELECT
    *
FROM
    tournament;

-- name: InsertDivision :exec
INSERT INTO divisions (series_id, name, ordering)
    VALUES (sqlc.arg (series_id), sqlc.arg (name), sqlc.arg (ordering));

-- name: InsertCountry :exec
INSERT INTO countries (country_id, name, abbreviation, flag_file)
    VALUES (sqlc.arg (country_id), sqlc.arg (name), sqlc.arg (abbreviation), sqlc.arg (flag_file));

-- name: InsertLocation :exec
INSERT INTO locations (id, name)
    VALUES (sqlc.arg (id), sqlc.arg (name));

-- name: InsertReservation :exec
INSERT INTO reservations (id, location, fieldname, reservationgroup)
    VALUES (sqlc.arg (id), sqlc.arg (location), sqlc.arg (fieldname), sqlc.arg (reservationgroup));

-- name: InsertPool :exec
INSERT INTO pools (pool_id, name, ordering, visible, continuingpool, placementpool, played, series, type)
    VALUES (sqlc.arg (pool_id), sqlc.arg (name), sqlc.arg (ordering), sqlc.arg (visible), sqlc.arg (continuingpool), sqlc.arg (placementpool), sqlc.arg (played), sqlc.arg (series), sqlc.arg (type));

-- name: InsertTeam :exec
INSERT INTO teams (team_id, name, pool, rank, valid, series, country, abbreviation, final_standing, final_standing_calculated)
    VALUES (sqlc.arg (team_id), sqlc.arg (name), sqlc.arg (pool), sqlc.arg (rank), sqlc.arg (valid), sqlc.arg (series), sqlc.arg (country), sqlc.arg (abbreviation), sqlc.arg (final_standing), sqlc.arg (final_standing_calculated));

-- name: InsertPlayer :exec
INSERT INTO players (player_id, firstname, lastname, team, num, games_played)
    VALUES (sqlc.arg (player_id), sqlc.arg (firstname), sqlc.arg (lastname), sqlc.arg (team), sqlc.arg (num), sqlc.arg (games_played));

