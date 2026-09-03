-- name: InsertTournament :exec
INSERT INTO tournament (season_id, name, starttime, endtime, iscurrent, type, isinternational, isnationalteams, showspiritpointsonlyoncomplete, lockteamspiritonsubmit, use_season_points, hide_time_on_scoresheet, hometeammode, event_readonly, maintenance_mode, public_event, api_public, timezone, spiritmode, host, base_path, app_version, archived_at)
    VALUES (sqlc.arg (season_id), sqlc.arg (name), sqlc.arg (starttime), sqlc.arg (endtime), sqlc.arg (iscurrent), sqlc.arg (type), sqlc.arg (isinternational), sqlc.arg (isnationalteams), sqlc.arg (showspiritpointsonlyoncomplete), sqlc.arg (lockteamspiritonsubmit), sqlc.arg (use_season_points), sqlc.arg (hide_time_on_scoresheet), sqlc.arg (hometeammode), sqlc.arg (event_readonly), sqlc.arg (maintenance_mode), sqlc.arg (public_event), sqlc.arg (api_public), sqlc.arg (timezone), sqlc.arg (spiritmode), sqlc.arg (host), sqlc.arg (base_path), sqlc.arg (app_version), sqlc.arg (archived_at));

-- name: GetTournament :one
SELECT
    *
FROM
    tournament;

-- name: InsertDivision :exec
INSERT INTO divisions (series_id, name, ordering)
    VALUES (sqlc.arg (series_id), sqlc.arg (name), sqlc.arg (ordering));

-- name: ListDivisions :many
SELECT
    *
FROM
    divisions
ORDER BY
    series_id;

-- name: InsertCountry :exec
INSERT INTO countries (country_id, name, abbreviation, flag_file)
    VALUES (sqlc.arg (country_id), sqlc.arg (name), sqlc.arg (abbreviation), sqlc.arg (flag_file));

-- name: ListCountries :many
SELECT
    *
FROM
    countries
ORDER BY
    country_id;

-- name: InsertLocation :exec
INSERT INTO locations (id, name)
    VALUES (sqlc.arg (id), sqlc.arg (name));

-- name: ListLocations :many
SELECT
    *
FROM
    locations
ORDER BY
    id;

-- name: InsertReservation :exec
INSERT INTO reservations (id, location, fieldname, reservationgroup)
    VALUES (sqlc.arg (id), sqlc.arg (location), sqlc.arg (fieldname), sqlc.arg (reservationgroup));

-- name: ListReservations :many
SELECT
    *
FROM
    reservations
ORDER BY
    id;

-- name: InsertPool :exec
-- color..follower are only known once a game in this pool has been fetched
-- they come from that game detail's poolinfo, not the reference endpoint's own pools[])
-- absent for a pool with no games, e.g. an unused placeholder bracket pool
INSERT INTO pools (pool_id, name, ordering, visible, continuingpool, placementpool, played, series, type, drawsallowed, playoff_template, color, timeslot, isfollower, teams, mvgames, timeoutlen, halftime, winningscore, timecap, scorecap, addscore, halftimescore, timeouts, timeoutsper, timeoutsovertime, timeoutstimecap, betweenpointslen, forfeitscore, forfeitagainst, follower)
    VALUES (sqlc.arg (pool_id), sqlc.arg (name), sqlc.arg (ordering), sqlc.arg (visible), sqlc.arg (continuingpool), sqlc.arg (placementpool), sqlc.arg (played), sqlc.arg (series), sqlc.arg (type), sqlc.arg (drawsallowed), sqlc.arg (playoff_template), sqlc.arg (color), sqlc.arg (timeslot), sqlc.arg (isfollower), sqlc.arg (teams), sqlc.arg (mvgames), sqlc.arg (timeoutlen), sqlc.arg (halftime), sqlc.arg (winningscore), sqlc.arg (timecap), sqlc.arg (scorecap), sqlc.arg (addscore), sqlc.arg (halftimescore), sqlc.arg (timeouts), sqlc.arg (timeoutsper), sqlc.arg (timeoutsovertime), sqlc.arg (timeoutstimecap), sqlc.arg (betweenpointslen), sqlc.arg (forfeitscore), sqlc.arg (forfeitagainst), sqlc.arg (follower));

-- name: ListPools :many
SELECT
    *
FROM
    pools
ORDER BY
    pool_id;

-- name: InsertTeam :exec
INSERT INTO teams (team_id, name, pool, rank, valid, series, country, abbreviation, final_standing, final_standing_calculated, club, clubname)
    VALUES (sqlc.arg (team_id), sqlc.arg (name), sqlc.arg (pool), sqlc.arg (rank), sqlc.arg (valid), sqlc.arg (series), sqlc.arg (country), sqlc.arg (abbreviation), sqlc.arg (final_standing), sqlc.arg (final_standing_calculated), sqlc.arg (club), sqlc.arg (clubname));

-- name: ListTeams :many
SELECT
    *
FROM
    teams
ORDER BY
    team_id;

-- name: InsertSchedulingName :exec
INSERT INTO scheduling_names (scheduling_id, name, frompool)
    VALUES (sqlc.arg (scheduling_id), sqlc.arg (name), sqlc.arg (frompool));

-- name: ListSchedulingNames :many
SELECT
    *
FROM
    scheduling_names
ORDER BY
    scheduling_id;

-- name: InsertPlayer :exec
INSERT INTO players (player_id, firstname, lastname, team, num, games_played)
    VALUES (sqlc.arg (player_id), sqlc.arg (firstname), sqlc.arg (lastname), sqlc.arg (team), sqlc.arg (num), sqlc.arg (games_played));

-- name: ListPlayers :many
SELECT
    *
FROM
    players
ORDER BY
    player_id;

-- name: InsertGame :exec
INSERT INTO games (game_id, hometeam, visitorteam, homescore, visitorscore, reservation, time, valid, halftime, official, respteam, resppers, isongoing, scheduling_name_home, scheduling_name_visitor, name, timeslot, homedefenses, visitordefenses, islive, liveurl, hasstarted, show_spirit, timer_start, timer_pause_start, timer_paused_duration, forfeit)
    VALUES (sqlc.arg (game_id), sqlc.arg (hometeam), sqlc.arg (visitorteam), sqlc.arg (homescore), sqlc.arg (visitorscore), sqlc.arg (reservation), sqlc.arg (time), sqlc.arg (valid), sqlc.arg (halftime), sqlc.arg (official), sqlc.arg (respteam), sqlc.arg (resppers), sqlc.arg (isongoing), sqlc.arg (scheduling_name_home), sqlc.arg (scheduling_name_visitor), sqlc.arg (name), sqlc.arg (timeslot), sqlc.arg (homedefenses), sqlc.arg (visitordefenses), sqlc.arg (islive), sqlc.arg (liveurl), sqlc.arg (hasstarted), sqlc.arg (show_spirit), sqlc.arg (timer_start), sqlc.arg (timer_pause_start), sqlc.arg (timer_paused_duration), sqlc.arg (forfeit));

-- name: ListGames :many
SELECT
    *
FROM
    games
ORDER BY
    game_id;

-- name: InsertGamePool :exec
INSERT INTO game_pools (game_id, pool_id, timetable)
    VALUES (sqlc.arg (game_id), sqlc.arg (pool_id), sqlc.arg (timetable));

-- name: ListGamePools :many
SELECT
    *
FROM
    game_pools
ORDER BY
    game_id,
    pool_id;

-- name: InsertGoal :exec
INSERT INTO goals (game_id, num, assist, scorer, time, homescore, visitorscore, ishomegoal, iscallahan, timestamp)
    VALUES (sqlc.arg (game_id), sqlc.arg (num), sqlc.arg (assist), sqlc.arg (scorer), sqlc.arg (time), sqlc.arg (homescore), sqlc.arg (visitorscore), sqlc.arg (ishomegoal), sqlc.arg (iscallahan), sqlc.arg (timestamp));

-- name: ListGoals :many
SELECT
    *
FROM
    goals
ORDER BY
    game_id,
    num;

-- name: InsertSpiritCategory :exec
INSERT INTO spirit_categories (category_id, mode, category_group, ordering, min, max, factor, label)
    VALUES (sqlc.arg (category_id), sqlc.arg (mode), sqlc.arg (category_group), sqlc.arg (ordering), sqlc.arg (min), sqlc.arg (max), sqlc.arg (factor), sqlc.arg (label));

-- name: ListSpiritCategories :many
SELECT
    *
FROM
    spirit_categories
ORDER BY
    category_id;

-- name: InsertSpiritScore :exec
INSERT INTO spirit_scores (game_id, team_id, category_id, value)
    VALUES (sqlc.arg (game_id), sqlc.arg (team_id), sqlc.arg (category_id), sqlc.arg (value));

-- name: ListSpiritScores :many
SELECT
    *
FROM
    spirit_scores
ORDER BY
    game_id,
    team_id,
    category_id;

-- name: InsertSpiritComment :exec
INSERT INTO spirit_comments (game_id, team_id, comment)
    VALUES (sqlc.arg (game_id), sqlc.arg (team_id), sqlc.arg (comment));

-- name: ListSpiritComments :many
SELECT
    *
FROM
    spirit_comments
ORDER BY
    game_id,
    team_id;

-- name: InsertPoolPlacement :exec
INSERT INTO pool_placements (pool_id, team_id, placement)
    VALUES (sqlc.arg (pool_id), sqlc.arg (team_id), sqlc.arg (placement));

-- name: ListPoolPlacements :many
SELECT
    *
FROM
    pool_placements
ORDER BY
    pool_id,
    team_id;

