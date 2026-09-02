-- Schema for a single tournament archived from a Live! by BULA 3.0.6 deployment.
--
-- Identical to v01_09_14's schema: 3.0.6 adds teams[].reg_id and games[].time_utc and
-- removes seed.sotg_token, but none of those touch this schema. reg_id is just the
-- organizer's external registration-system id for the team, not meaningful archival
-- data; time_utc is fully reconstructible from games.time + tournament.timezone; and
-- sotg_token (a spirit-submission token, not a result) was never stored to begin with.
--
-- Every table mirrors the underlying UltiOrganizer table it's drawn from
-- with the original MySQL type noted per column.
-- Columns not exposed via API are kept, nullable, rather than dropped
-- prune once confirmed unreachable via the API
-- UltiOrganizer's own schema has no foreign keys, these have been inferred
--
-- Nothing derivable from other stored columns is stored redundantly
-- e.g. a game's status or a team's win/loss record
-- The exceptions are values with no other source in this archive at all
-- e.g. final_standing override, player's games_played
-- those are kept as reported, commented with why.
--
-- uo_country
CREATE TABLE countries (
    country_id integer PRIMARY KEY, -- int(10)
    name text NOT NULL, -- varchar(50)
    abbreviation text, -- char(3)
    flag_file text, -- varchar(50)
    valid integer NOT NULL DEFAULT 1 -- tinyint(4)
);

-- uo_location
-- API's Reservation.name (venue name) is this table's `name`
-- joined via uo_reservation.location
-- uo_reservation has no name column of its own.
CREATE TABLE locations (
    id integer PRIMARY KEY, -- int(10)
    name text NOT NULL, -- varchar(50)
    fields integer NOT NULL DEFAULT 1, -- int(5)
    indoor integer NOT NULL DEFAULT 0, -- tinyint(1)
    address text, -- varchar(255)
    lat real, -- float(17,13)
    lng real -- float(17,13)
    -- info_fi_FI_utf8/info_en_GB_utf8 (localized info text) skipped: not exposed via the API at all
);

-- uo_season
-- plus archive-specific metadata that isn't from UltiOrganizer at all.
CREATE TABLE tournament (
    season_id text NOT NULL, -- varchar(10)
    name text, -- varchar(50)
    starttime text, -- datetime
    endtime text, -- datetime
    iscurrent integer NOT NULL DEFAULT 0, -- tinyint(4)
    enrollopen integer NOT NULL DEFAULT 0, -- tinyint(1)
    enroll_deadline text, -- datetime
    type text, -- varchar(20)
    istournament integer DEFAULT 0, -- tinyint(1)
    isinternational integer DEFAULT 0, -- tinyint(1)
    isnationalteams integer DEFAULT 0, -- tinyint(1)
    organizer text, -- varchar(50)
    category text, -- varchar(50)
    spiritpoints integer, -- tinyint(1)
    showspiritpoints integer DEFAULT 0, -- tinyint(1)
    timezone text, -- varchar(50)
    -- archive metadata: not from uo_season
    host text NOT NULL,
    base_path text NOT NULL,
    app_version text,
    archived_at text NOT NULL
);

-- uo_series
CREATE TABLE divisions (
    series_id integer PRIMARY KEY, -- int(10)
    name text, -- varchar(50)
    ordering text, -- varchar(1)
    season text, -- varchar(50); redundant with the single tournament
    valid integer NOT NULL DEFAULT 0, -- tinyint(4)
    type text, -- varchar(20)
    color text, -- varchar(6)
    pool_template integer -- int(10)
);

-- uo_pool
CREATE TABLE pools (
    pool_id integer PRIMARY KEY, -- int(10)
    name text, -- varchar(50)
    ordering text, -- varchar(20)
    visible integer NOT NULL, -- tinyint(1)
    continuingpool integer NOT NULL, -- tinyint(1)
    placementpool integer DEFAULT 0, -- tinyint(1)
    teams integer, -- int(10); configured team count
    mvgames integer, -- int(10)
    timeoutlen integer, -- int(10)
    halftime integer, -- int(10)
    winningscore integer, -- smallint(5)
    timecap integer, -- int(10)
    scorecap integer, -- smallint(5)
    played integer NOT NULL, -- tinyint(1)
    addscore integer, -- int(10)
    halftimescore integer, -- int(10)
    timeouts integer, -- int(10)
    timeoutsper text, -- varchar(5)
    timeoutsovertime integer, -- int(10)
    timeoutstimecap text, -- varchar(5)
    betweenpointslen integer, -- int(10)
    series integer REFERENCES divisions (series_id), -- int(10)
    type integer NOT NULL DEFAULT 1, -- int(10)
    timeslot integer, -- int(10)
    color text, -- varchar(6)
    forfeitscore integer, -- int(10)
    forfeitagainst integer, -- int(10)
    follower integer -- int(10)
);

-- uo_reservation
CREATE TABLE reservations (
    id integer PRIMARY KEY, -- int(10)
    location integer NOT NULL REFERENCES locations (id), -- int(10)
    fieldname text, -- varchar(50)
    reservationgroup text, -- varchar(50)
    starttime text, -- datetime
    endtime text, -- datetime
    season text, -- varchar(10)
    timeslots text, -- varchar(100)
    date text -- datetime
);

-- uo_team
CREATE TABLE teams (
    team_id integer PRIMARY KEY, -- int(10)
    name text, -- varchar(50)
    pool integer REFERENCES pools (pool_id), -- int(10)
    rank integer, -- smallint(5); seed going into the event
    activerank integer, -- int(10)
    valid integer NOT NULL, -- tinyint(1)
    series integer REFERENCES divisions (series_id), -- int(10)
    country integer REFERENCES countries (country_id), -- int(10)
    abbreviation text, -- varchar(15)
    -- final_standing/final_standing_calculated kept as reported
    -- final_standing can be an organizer override, and *_calculated reflects
    -- bracket-resolution logic this archive doesn't model
    final_standing integer,
    final_standing_calculated integer,
    club_name text -- bare club_name, not an id
);

-- uo_player
CREATE TABLE players (
    player_id integer PRIMARY KEY, -- int(10)
    firstname text, -- varchar(40) DEFAULT ''
    lastname text, -- varchar(40) DEFAULT ''
    team integer REFERENCES teams (team_id), -- int(10)
    num integer, -- tinyint(3) unsigned
    accreditation_id text, -- varchar(150)
    accredited integer NOT NULL DEFAULT 0, -- tinyint(1)
    profile_id integer, -- int(10)
    -- games_played kept as reported
    -- player can play a game recording zero goals or assists
    -- so it isn't derivable from the goals table alone
    -- goals/assists/callahans are NOT kept here they are recorded in goals table
    games_played integer
);

-- uo_game
CREATE TABLE games (
    game_id integer PRIMARY KEY, -- int(10)
    hometeam integer REFERENCES teams (team_id), -- int(10)
    visitorteam integer REFERENCES teams (team_id), -- int(10)
    homescore integer, -- smallint(5); absent (not just null) for a scheduled game
    visitorscore integer, -- smallint(5)
    reservation integer REFERENCES reservations (id), -- int(10)
    time text, -- datetime; absent for an unscheduled game
    pool integer REFERENCES pools (pool_id), -- int(10)
    valid integer NOT NULL, -- tinyint(1)
    halftime integer, -- int(10)
    official text, -- varchar(50)
    respteam integer REFERENCES teams (team_id), -- int(10)
    resppers integer, -- int(10)
    homesotg integer, -- int(10)
    visitorsotg integer, -- int(10)
    isongoing integer DEFAULT 0, -- tinyint(1)
    scheduling_name_home integer, -- int(10)
    scheduling_name_visitor integer, -- int(10)
    name integer, -- int(10); scheduling-name id, resolved via uo_scheduling_name (not modeled)
    timeslot integer, -- int(10)
    defenses_total integer, -- smallint(5)
    homedefenses integer, -- smallint(5)
    visitordefenses integer, -- smallint(5)
    islive integer DEFAULT 0, -- tinyint(1)
    liveurl text -- varchar(255)
);

-- uo_goal
CREATE TABLE goals (
    game_id integer NOT NULL REFERENCES games (game_id), -- int(10)
    num integer NOT NULL, -- smallint(5)
    assist integer REFERENCES players (player_id), -- int(10)
    scorer integer REFERENCES players (player_id), -- int(10)
    time integer, -- smallint(5); seconds from the start of the game
    homescore integer, -- tinyint(3) unsigned; running score after this goal
    visitorscore integer, -- tinyint(3) unsigned
    ishomegoal integer NOT NULL, -- tinyint(1)
    iscallahan integer NOT NULL, -- tinyint(1)
    PRIMARY KEY (game_id, num)
);

-- uo_spirit
-- One row per (game, recipient team)
-- team_id is for the recipient of this score
CREATE TABLE spirit_scores (
    game_id integer NOT NULL REFERENCES games (game_id), -- int(10)
    team_id integer NOT NULL REFERENCES teams (team_id), -- int(10); the team this score is for
    cat1 integer NOT NULL DEFAULT 0, -- tinyint(2)
    cat2 integer NOT NULL DEFAULT 0, -- tinyint(2)
    cat3 integer NOT NULL DEFAULT 0, -- tinyint(2)
    cat4 integer NOT NULL DEFAULT 0, -- tinyint(2)
    cat5 integer NOT NULL DEFAULT 0, -- tinyint(2)
    comments text, -- TEXT DEFAULT NULL
    PRIMARY KEY (game_id, team_id)
);

