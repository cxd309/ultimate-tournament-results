-- name: InsertTournament :exec
INSERT INTO tournament (season_id, name, starttime, endtime, timezone, spiritmode, host, base_path, app_version, archived_at)
    VALUES (sqlc.arg (season_id), sqlc.arg (name), sqlc.arg (starttime), sqlc.arg (endtime), sqlc.arg (timezone), sqlc.arg (spiritmode), sqlc.arg (host), sqlc.arg (base_path), sqlc.arg (app_version), sqlc.arg (archived_at));

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
INSERT INTO pools (pool_id, name, ordering, visible, continuingpool, placementpool, played, series, type, drawsallowed, playoff_template)
    VALUES (sqlc.arg (pool_id), sqlc.arg (name), sqlc.arg (ordering), sqlc.arg (visible), sqlc.arg (continuingpool), sqlc.arg (placementpool), sqlc.arg (played), sqlc.arg (series), sqlc.arg (type), sqlc.arg (drawsallowed), sqlc.arg (playoff_template));

-- name: InsertTeam :exec
INSERT INTO teams (team_id, name, pool, rank, valid, series, country, abbreviation, final_standing, final_standing_calculated, club_name)
    VALUES (sqlc.arg (team_id), sqlc.arg (name), sqlc.arg (pool), sqlc.arg (rank), sqlc.arg (valid), sqlc.arg (series), sqlc.arg (country), sqlc.arg (abbreviation), sqlc.arg (final_standing), sqlc.arg (final_standing_calculated), sqlc.arg (club_name));

-- name: InsertPlayer :exec
INSERT INTO players (player_id, firstname, lastname, team, num, games_played)
    VALUES (sqlc.arg (player_id), sqlc.arg (firstname), sqlc.arg (lastname), sqlc.arg (team), sqlc.arg (num), sqlc.arg (games_played));

-- name: InsertGame :exec
INSERT INTO games (game_id, hometeam, visitorteam, homescore, visitorscore, reservation, time, valid, halftime, official, respteam, resppers, isongoing, scheduling_name_home, scheduling_name_visitor, name, timeslot, homedefenses, visitordefenses, islive, liveurl, hasstarted, show_spirit, timer_start, timer_pause_start, timer_paused_duration, forfeit)
    VALUES (sqlc.arg (game_id), sqlc.arg (hometeam), sqlc.arg (visitorteam), sqlc.arg (homescore), sqlc.arg (visitorscore), sqlc.arg (reservation), sqlc.arg (time), sqlc.arg (valid), sqlc.arg (halftime), sqlc.arg (official), sqlc.arg (respteam), sqlc.arg (resppers), sqlc.arg (isongoing), sqlc.arg (scheduling_name_home), sqlc.arg (scheduling_name_visitor), sqlc.arg (name), sqlc.arg (timeslot), sqlc.arg (homedefenses), sqlc.arg (visitordefenses), sqlc.arg (islive), sqlc.arg (liveurl), sqlc.arg (hasstarted), sqlc.arg (show_spirit), sqlc.arg (timer_start), sqlc.arg (timer_pause_start), sqlc.arg (timer_paused_duration), sqlc.arg (forfeit));

-- name: InsertGamePool :exec
INSERT INTO game_pools (game_id, pool_id, timetable)
    VALUES (sqlc.arg (game_id), sqlc.arg (pool_id), sqlc.arg (timetable));

-- name: InsertGoal :exec
INSERT INTO goals (game_id, num, assist, scorer, time, homescore, visitorscore, ishomegoal, iscallahan, timestamp)
    VALUES (sqlc.arg (game_id), sqlc.arg (num), sqlc.arg (assist), sqlc.arg (scorer), sqlc.arg (time), sqlc.arg (homescore), sqlc.arg (visitorscore), sqlc.arg (ishomegoal), sqlc.arg (iscallahan), sqlc.arg (timestamp));

-- name: InsertSpiritCategory :exec
INSERT INTO spirit_categories (category_id, mode, category_group, ordering, min, max, factor, label)
    VALUES (sqlc.arg (category_id), sqlc.arg (mode), sqlc.arg (category_group), sqlc.arg (ordering), sqlc.arg (min), sqlc.arg (max), sqlc.arg (factor), sqlc.arg (label));

-- name: InsertSpiritScore :exec
INSERT INTO spirit_scores (game_id, team_id, category_id, value)
    VALUES (sqlc.arg (game_id), sqlc.arg (team_id), sqlc.arg (category_id), sqlc.arg (value));

-- name: InsertSpiritComment :exec
INSERT INTO spirit_comments (game_id, team_id, comment)
    VALUES (sqlc.arg (game_id), sqlc.arg (team_id), sqlc.arg (comment));

-- name: InsertPoolPlacement :exec
INSERT INTO pool_placements (pool_id, team_id, placement)
    VALUES (sqlc.arg (pool_id), sqlc.arg (team_id), sqlc.arg (placement));

