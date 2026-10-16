-- Schema extension for a single tournament archived from a classic UltiOrganizer site
-- (one with no Live! by BULA deployment), by scraping its public HTML pages and CSV exports
--
-- Applied on top of db/v01_09_14/schema.sql, which is left exactly as it is:
-- every shared table keeps the same columns, so the 1.9.14-1.9.17 store and publisher
-- read a uo_01 archive unchanged and docs/archive/ gets the same Live! JSON
--
-- The tables below hold what UltiOrganizer shows publicly that the Live! schema has
-- no place for. Each mirrors the UltiOrganizer table it's drawn from with the original
-- MySQL type noted per column; columns the public pages never show are left out
-- rather than kept as always-null, and noted per table
--
-- uo_played
-- one row per player listed on a game's scoreboard, i.e. everyone entered as having
-- played that game, including those who recorded no goals or assists
-- accredited/acknowledged skipped: never shown publicly
CREATE TABLE played (
    game_id integer NOT NULL REFERENCES games (game_id), -- int(10)
    player_id integer NOT NULL REFERENCES players (player_id), -- int(10)
    num integer, -- smallint(5); shirt number worn in this game
    captain integer NOT NULL DEFAULT 0, -- tinyint(1); shown as "(C)" on the game page
    PRIMARY KEY (game_id, player_id)
);

-- uo_gameevent
-- timeouts, turnovers and starting offence as listed in a game's scoresheet
-- the game page only lists an event in the row of the goal that follows it, so an event
-- after the last goal, or a timeout at 0:00 or 1:00 (which the page deliberately hides),
-- is not recoverable
-- num is this archive's own 1-based display order, not UltiOrganizer's num
-- info skipped: never shown publicly
CREATE TABLE game_events (
    game_id integer NOT NULL REFERENCES games (game_id), -- int(10)
    num integer NOT NULL, -- smallint(5)
    time integer, -- smallint(5); seconds from the start of the game
    type text NOT NULL, -- varchar(10); timeout, turnover or offence
    ishome integer NOT NULL, -- tinyint(1)
    PRIMARY KEY (game_id, num)
);

-- uo_team_pool
-- every pool a team was placed in, not only teams.pool
-- activerank is the "#" standings column of a round-robin pool page; null for other
-- pool types, which don't print it
-- rank skipped: never shown publicly
CREATE TABLE team_pools (
    team_id integer NOT NULL REFERENCES teams (team_id), -- int(10)
    pool_id integer NOT NULL REFERENCES pools (pool_id), -- int(10)
    activerank integer, -- int(10)
    PRIMARY KEY (team_id, pool_id)
);

-- uo_game_pool
-- every pool page a game is listed on; a game can count towards more than one pool
-- (e.g. a crossover), games.pool is only the one it's scheduled under
-- timetable skipped: never shown publicly
CREATE TABLE game_pools (
    game_id integer NOT NULL REFERENCES games (game_id), -- int(10)
    pool_id integer NOT NULL REFERENCES pools (pool_id), -- int(10)
    PRIMARY KEY (game_id, pool_id)
);

-- uo_moveteams
-- which placing in a round-robin pool continues to which pool, from the pool page's
-- colour-coded standings and its "Continues to" list
-- torank/ismoved/scheduling_id skipped: never shown publicly
CREATE TABLE move_teams (
    frompool integer NOT NULL REFERENCES pools (pool_id), -- int(10)
    fromplacing integer NOT NULL, -- int(5)
    topool integer NOT NULL REFERENCES pools (pool_id), -- int(10)
    PRIMARY KEY (frompool, fromplacing)
);

-- uo_club
CREATE TABLE clubs (
    club_id integer PRIMARY KEY, -- int(10)
    name text NOT NULL -- varchar(50)
);

-- uo_team.club
-- kept out of teams so the shared teams table stays column-for-column v01_09_14
CREATE TABLE team_clubs (
    team_id integer PRIMARY KEY REFERENCES teams (team_id), -- int(10)
    club_id integer NOT NULL REFERENCES clubs (club_id) -- int(10)
);

-- uo_team_profile
-- only for teams whose card shows any of these
-- image/ffindr_id skipped: never shown publicly
CREATE TABLE team_profiles (
    team_id integer PRIMARY KEY REFERENCES teams (team_id), -- int(11)
    coach text, -- varchar(100)
    captain text, -- varchar(100)
    story text, -- text
    achievements text, -- text
    profile_image text -- varchar(20); filename only, the image itself isn't archived
);

-- uo_urls
-- links shown on game and team pages: "Team pages" (ismedialink 0) and
-- "Photos and Videos" (ismedialink 1)
-- player links skipped: they're only on each player's own card, one request per player
-- url_id/ordering/publisher_id skipped: never shown publicly
CREATE TABLE urls (
    owner text NOT NULL, -- varchar(15); game or team
    owner_id integer NOT NULL, -- varchar(15)
    type text, -- varchar(15); link icon name, e.g. youtube
    name text, -- varchar(150)
    url text NOT NULL, -- varchar(500)
    ismedialink integer NOT NULL DEFAULT 0, -- tinyint(1)
    mediaowner text -- varchar(100)
);

