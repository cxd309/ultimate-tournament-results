-- name: InsertTournament :exec
INSERT INTO tournament (season_id, name, starttime, endtime, iscurrent, type, isinternational, isnationalteams, timezone, host, base_path, app_version, archived_at)
    VALUES (sqlc.arg (season_id), sqlc.arg (name), sqlc.arg (starttime), sqlc.arg (endtime), sqlc.arg (iscurrent), sqlc.arg (type), sqlc.arg (isinternational), sqlc.arg (isnationalteams), sqlc.arg (timezone), sqlc.arg (host), sqlc.arg (base_path), sqlc.arg (app_version), sqlc.arg (archived_at));

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
-- teams to follower are only known once a game in this pool has been fetched
-- (they come from that game detail's poolinfo, not the reference endpoint's own pools[])
-- absent for a pool with no games, e.g. an unused placeholder bracket pool
INSERT INTO pools (pool_id, name, ordering, visible, continuingpool, placementpool, played, series, type, color, timeslot, teams, mvgames, timeoutlen, halftime, winningscore, timecap, scorecap, addscore, halftimescore, timeouts, timeoutsper, timeoutsovertime, timeoutstimecap, betweenpointslen, forfeitscore, forfeitagainst, follower)
    VALUES (sqlc.arg (pool_id), sqlc.arg (name), sqlc.arg (ordering), sqlc.arg (visible), sqlc.arg (continuingpool), sqlc.arg (placementpool), sqlc.arg (played), sqlc.arg (series), sqlc.arg (type), sqlc.arg (color), sqlc.arg (timeslot), sqlc.arg (teams), sqlc.arg (mvgames), sqlc.arg (timeoutlen), sqlc.arg (halftime), sqlc.arg (winningscore), sqlc.arg (timecap), sqlc.arg (scorecap), sqlc.arg (addscore), sqlc.arg (halftimescore), sqlc.arg (timeouts), sqlc.arg (timeoutsper), sqlc.arg (timeoutsovertime), sqlc.arg (timeoutstimecap), sqlc.arg (betweenpointslen), sqlc.arg (forfeitscore), sqlc.arg (forfeitagainst), sqlc.arg (follower));

-- name: InsertTeam :exec
INSERT INTO teams (team_id, name, pool, rank, valid, series, country, abbreviation, final_standing, final_standing_calculated, club_name)
    VALUES (sqlc.arg (team_id), sqlc.arg (name), sqlc.arg (pool), sqlc.arg (rank), sqlc.arg (valid), sqlc.arg (series), sqlc.arg (country), sqlc.arg (abbreviation), sqlc.arg (final_standing), sqlc.arg (final_standing_calculated), sqlc.arg (club_name));

-- name: InsertPlayer :exec
INSERT INTO players (player_id, firstname, lastname, team, num, games_played)
    VALUES (sqlc.arg (player_id), sqlc.arg (firstname), sqlc.arg (lastname), sqlc.arg (team), sqlc.arg (num), sqlc.arg (games_played));

-- name: InsertSchedulingName :exec
INSERT INTO scheduling_names (scheduling_id, name)
    VALUES (sqlc.arg (scheduling_id), sqlc.arg (name));

-- name: InsertGame :exec
INSERT INTO games (game_id, hometeam, visitorteam, homescore, visitorscore, reservation, time, pool, valid, halftime, official, respteam, resppers, homesotg, visitorsotg, isongoing, scheduling_name_home, scheduling_name_visitor, name, timeslot, homedefenses, visitordefenses, islive, liveurl)
    VALUES (sqlc.arg (game_id), sqlc.arg (hometeam), sqlc.arg (visitorteam), sqlc.arg (homescore), sqlc.arg (visitorscore), sqlc.arg (reservation), sqlc.arg (time), sqlc.arg (pool), sqlc.arg (valid), sqlc.arg (halftime), sqlc.arg (official), sqlc.arg (respteam), sqlc.arg (resppers), sqlc.arg (homesotg), sqlc.arg (visitorsotg), sqlc.arg (isongoing), sqlc.arg (scheduling_name_home), sqlc.arg (scheduling_name_visitor), sqlc.arg (name), sqlc.arg (timeslot), sqlc.arg (homedefenses), sqlc.arg (visitordefenses), sqlc.arg (islive), sqlc.arg (liveurl));

-- name: InsertGoal :exec
INSERT INTO goals (game_id, num, assist, scorer, time, homescore, visitorscore, ishomegoal, iscallahan)
    VALUES (sqlc.arg (game_id), sqlc.arg (num), sqlc.arg (assist), sqlc.arg (scorer), sqlc.arg (time), sqlc.arg (homescore), sqlc.arg (visitorscore), sqlc.arg (ishomegoal), sqlc.arg (iscallahan));

-- name: InsertSpiritScore :exec
INSERT INTO spirit_scores (game_id, team_id, cat1, cat2, cat3, cat4, cat5, comments)
    VALUES (sqlc.arg (game_id), sqlc.arg (team_id), sqlc.arg (cat1), sqlc.arg (cat2), sqlc.arg (cat3), sqlc.arg (cat4), sqlc.arg (cat5), sqlc.arg (comments));

