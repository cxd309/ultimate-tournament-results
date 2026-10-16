-- uo_01 only adds queries the v01_09_14 store has no equivalent for: every other
-- shared table is written through store/v01_09_14 unchanged
--
-- name: InsertLocation :exec
-- the v01_09_14 insert only takes id/name, the Live! API exposes nothing more
INSERT INTO locations (id, name, fields, indoor, address, lat, lng)
    VALUES (sqlc.arg (id), sqlc.arg (name), sqlc.arg (fields), sqlc.arg (indoor), sqlc.arg (address), sqlc.arg (lat), sqlc.arg (lng));

-- name: InsertReservation :exec
-- the v01_09_14 insert has no times, the Live! API exposes none
INSERT INTO reservations (id, location, fieldname, reservationgroup, starttime, endtime, season, date)
    VALUES (sqlc.arg (id), sqlc.arg (location), sqlc.arg (fieldname), sqlc.arg (reservationgroup), sqlc.arg (starttime), sqlc.arg (endtime), sqlc.arg (season), sqlc.arg (date));

-- name: InsertPlayed :exec
INSERT INTO played (game_id, player_id, num, captain)
    VALUES (sqlc.arg (game_id), sqlc.arg (player_id), sqlc.arg (num), sqlc.arg (captain));

-- name: ListPlayed :many
SELECT
    *
FROM
    played
ORDER BY
    game_id,
    player_id;

-- name: InsertGameEvent :exec
INSERT INTO game_events (game_id, num, time, type, ishome)
    VALUES (sqlc.arg (game_id), sqlc.arg (num), sqlc.arg (time), sqlc.arg (type), sqlc.arg (ishome));

-- name: ListGameEvents :many
SELECT
    *
FROM
    game_events
ORDER BY
    game_id,
    num;

-- name: InsertTeamPool :exec
INSERT INTO team_pools (team_id, pool_id, activerank)
    VALUES (sqlc.arg (team_id), sqlc.arg (pool_id), sqlc.arg (activerank));

-- name: ListTeamPools :many
SELECT
    *
FROM
    team_pools
ORDER BY
    team_id,
    pool_id;

-- name: InsertGamePool :exec
INSERT INTO game_pools (game_id, pool_id)
    VALUES (sqlc.arg (game_id), sqlc.arg (pool_id));

-- name: ListGamePools :many
SELECT
    *
FROM
    game_pools
ORDER BY
    game_id,
    pool_id;

-- name: InsertMoveTeam :exec
INSERT INTO move_teams (frompool, fromplacing, topool)
    VALUES (sqlc.arg (frompool), sqlc.arg (fromplacing), sqlc.arg (topool));

-- name: ListMoveTeams :many
SELECT
    *
FROM
    move_teams
ORDER BY
    frompool,
    fromplacing;

-- name: InsertClub :exec
INSERT INTO clubs (club_id, name)
    VALUES (sqlc.arg (club_id), sqlc.arg (name));

-- name: InsertTeamClub :exec
INSERT INTO team_clubs (team_id, club_id)
    VALUES (sqlc.arg (team_id), sqlc.arg (club_id));

-- name: InsertTeamProfile :exec
INSERT INTO team_profiles (team_id, coach, captain, story, achievements, profile_image)
    VALUES (sqlc.arg (team_id), sqlc.arg (coach), sqlc.arg (captain), sqlc.arg (story), sqlc.arg (achievements), sqlc.arg (profile_image));

-- name: InsertURL :exec
INSERT INTO urls (owner, owner_id, type, name, url, ismedialink, mediaowner)
    VALUES (sqlc.arg (owner), sqlc.arg (owner_id), sqlc.arg (type), sqlc.arg (name), sqlc.arg (url), sqlc.arg (ismedialink), sqlc.arg (mediaowner));

