-- Schema for a single tournament archived from a Live! by BULA 3.0.6 deployment
-- (UltiOrganizer 4).
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
-- UltiOrganizer 4 changes:
-- uo_game drops pool column -- games join pools only through uo_game_pool now
-- uo_spirt table dropped -- no fixed spirit categories
-- uo_spirit_category table created -- flexible spirit categories
-- uo_spirit_score table created -- a row for a score per game/team/category
-- uo_game drops homesotg/visitorsotg/defenses_total columns
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
    -- uo_location_info (localized info text, one row per locale on this line) skipped:
    -- not exposed via the API at all
);

-- uo_season
-- plus archive-specific metadata that isn't from UltiOrganizer at all.
-- reg_id not included, external registration system data not relevant
-- enrollopen/enroll_deadline/istournament/organizer/category are real columns too but
-- the reference endpoint strips all five and no other in-scope endpoint exposes them
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
    showspiritpoints integer DEFAULT 0, -- tinyint(1)
    showspiritcomments integer DEFAULT 0, -- tinyint(1)
    showspiritpointsonlyoncomplete integer DEFAULT 1, -- tinyint(1)
    lockteamspiritonsubmit integer DEFAULT 1, -- tinyint(1)
    use_season_points integer DEFAULT 0, -- tinyint(1)
    hide_time_on_scoresheet integer DEFAULT 0, -- tinyint(1)
    hometeammode integer DEFAULT 0, -- tinyint(1)
    event_readonly integer DEFAULT 0, -- tinyint(1)
    maintenance_mode integer DEFAULT 0, -- tinyint(1)
    public_event integer NOT NULL DEFAULT 0, -- tinyint(1)
    api_public integer DEFAULT 0, -- tinyint(1)
    timezone text, -- varchar(50)
    spiritmode integer, -- int(10); which spirit scoring system is in use
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
    follower integer, -- int(10)
    drawsallowed integer, -- smallint(5) DEFAULT 0
    playoff_template text, -- varchar(30)
    -- isfollower: computed by the reference endpoint from `follower`, not a raw column
    -- (`follower` itself, the target pool id, is only exposed via poolinfo, see above)
    isfollower integer NOT NULL DEFAULT 0 -- tinyint(1)
);

-- uo_reservation
CREATE TABLE reservations (
    id integer PRIMARY KEY, -- int(10)
    location integer REFERENCES locations (id), -- int(10)
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
    -- club is a real uo_team FK (nullable, null at a national-teams event); clubname is
    -- the resolved text, only available from the team-detail endpoint, not reference
    -- sotg_token (spirit-submission token) not modeled: a credential, not a result
    club integer, -- int(10)
    clubname text -- varchar(50)
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
    -- a player can play a game recording zero goals or assists
    -- so it isn't derivable from the goals table alone
    -- goals/assists/callahans are NOT kept here they are recorded in goals table
    games_played integer
);

-- uo_scheduling_name
-- resolves a scheduling-name id (games.name/scheduling_name_home/scheduling_name_visitor)
-- to its display text: a bracket-slot placeholder label like "4A", or a fixture's own
-- short name like "wxo1"
-- frompool (uo_moveteams.frompool, new in UltiOrganizer 4) names the pool a placeholder
-- slot's team will be drawn from, when known.
CREATE TABLE scheduling_names (
    scheduling_id integer PRIMARY KEY, -- int(10)
    name text NOT NULL, -- varchar(100)
    frompool integer REFERENCES pools (pool_id) -- int(10)
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
    valid integer NOT NULL, -- tinyint(1)
    halftime integer, -- int(10)
    official text, -- varchar(50)
    respteam integer REFERENCES teams (team_id), -- int(10)
    resppers integer, -- int(10)
    isongoing integer DEFAULT 0, -- tinyint(1)
    scheduling_name_home integer REFERENCES scheduling_names (scheduling_id), -- int(10)
    scheduling_name_visitor integer REFERENCES scheduling_names (scheduling_id), -- int(10)
    name integer REFERENCES scheduling_names (scheduling_id), -- int(10); scheduling-name id
    timeslot integer, -- int(10)
    homedefenses integer, -- smallint(5) DEFAULT 0
    visitordefenses integer, -- smallint(5) DEFAULT 0
    islive integer DEFAULT 0, -- tinyint(1)
    liveurl text, -- varchar(512)
    hasstarted integer, -- tinyint(1) DEFAULT 0; started flag (can be 0, 1 or 2)
    show_spirit integer DEFAULT 0, -- tinyint(1); are spirit scores displayed
    timer_start integer, -- bigint(20); unix seconds the game clock was started
    timer_pause_start integer, -- bigint(20); unix seconds the clock was paused
    timer_paused_duration integer DEFAULT 0, -- bigint(20) NOT NULL; total seconds paused
    -- forfeit: whether the game was forfeited. A forfeit reports completed with a 0-0
    -- scoreline on this line, so this is the only way to distinguish it from a genuine
    -- 0-0 result.
    forfeit integer DEFAULT 0 -- tinyint(1) NOT NULL;
);

-- uo_game_pool
-- some games belong to multiple pools so this table manages the many-many mapping
-- e.g. power pools where the result from a previous pool is carried through
-- each game is owned (scheduled) by a single pool which is shown by timetable=1
-- a "borrowed" game has timetable=0
CREATE TABLE game_pools (
    game_id integer NOT NULL REFERENCES games (game_id), -- int(10); real column `game`
    pool_id integer NOT NULL REFERENCES pools (pool_id), -- int(10); real column `pool`
    timetable integer NOT NULL, -- tinyint(1); 1 = this is the pool that "owns" this game
    PRIMARY KEY (game_id, pool_id)
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
    timestamp text, -- datetime DEFAULT current_timestamp(); timestamp of goal being recorded
    PRIMARY KEY (game_id, num)
);

-- uo_spirit_category
-- Spirit category definitions per scoring mode
-- this will be set across a whole deployment for multiple tournaments same as countries
-- for this tool there should only be a single mode as this is for a single tournament
CREATE TABLE spirit_categories (
    category_id integer PRIMARY KEY, -- int(10)
    mode integer NOT NULL, -- int(10); scoring scheme id, matches tournament.spiritmode
    category_group integer NOT NULL DEFAULT 1, -- int(5); real column `group`
    -- ordering:
    -- position in the category list, counting from 1; real column `index`.
    -- Also the digits in the API's "catN" key for this category.
    ordering integer NOT NULL, -- int(5)
    min integer NOT NULL DEFAULT 0, -- int(5)
    max integer NOT NULL DEFAULT 4, -- int(5)
    factor integer NOT NULL DEFAULT 1, -- int(5)
    label text NOT NULL -- text; real column `text`
);

-- uo_spirit_score
-- a single spirit score in a category for a team in a game
-- team_id is the who the score is FOR
-- who gave it can be assumed to be the other team in the game
CREATE TABLE spirit_scores (
    game_id integer NOT NULL REFERENCES games (game_id), -- int(10)
    team_id integer NOT NULL REFERENCES teams (team_id), -- int(10); the team recieving this score
    category_id integer NOT NULL REFERENCES spirit_categories (category_id), -- int(10)
    value integer, -- int(3); 0 is a real submitted score, null means not yet visible
    PRIMARY KEY (game_id, team_id, category_id)
);

-- uo_comment
-- uo_comment is a generic comment table for comments across UltiOrganizer
-- this doesn't replicate the whole shape, it is only focused on spirit comments
CREATE TABLE spirit_comments (
    game_id integer NOT NULL REFERENCES games (game_id),
    team_id integer NOT NULL REFERENCES teams (team_id),
    comment text,
    PRIMARY KEY (game_id, team_id)
);

-- UltiOrganizer computes this with standings logic and does not cache result
-- Same as teams.final_standing_calculated
-- This archive does not include standings logic so just has a specific table
CREATE TABLE pool_placements (
    pool_id integer NOT NULL REFERENCES pools (pool_id),
    team_id integer NOT NULL REFERENCES teams (team_id),
    placement integer, -- null while the pool has no resolved rank for that team yet
    PRIMARY KEY (pool_id, team_id)
);

