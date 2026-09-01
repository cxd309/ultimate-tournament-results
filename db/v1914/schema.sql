-- Schema for a single tournament
-- archived from a Live! by BULA 1.9.14-1.9.16 deployment
-- openapi-1.9.14.yaml in the live-by-bula-openapi repo
CREATE TABLE tournament (
    -- no primary key: exactly one row, written once by one process into a fresh file
    event_name text NOT NULL,
    host text NOT NULL,
    season_id text NOT NULL,
    base_path text NOT NULL,
    app_version text, -- raw app_version string as reported
    start_date text,
    end_date text,
    timezone text,
    status text, -- season.status at archive time
    archived_at text NOT NULL
);

CREATE TABLE divisions (
    id integer PRIMARY KEY,
    series_id integer NOT NULL UNIQUE, -- external id
    name text NOT NULL,
    ordering text
);

CREATE TABLE pools (
    id integer PRIMARY KEY,
    pool_id integer NOT NULL UNIQUE, -- external id
    division_id integer REFERENCES divisions (id),
    name text NOT NULL,
    ordering text,
    pool_type integer
);

CREATE TABLE countries (
    id integer PRIMARY KEY,
    country_ext_id integer NOT NULL UNIQUE, -- stable across events per the spec
    name text NOT NULL,
    abbreviation text,
    flag_file text
);

CREATE TABLE teams (
    id integer PRIMARY KEY,
    team_id integer NOT NULL UNIQUE, -- external id
    division_id integer REFERENCES divisions (id),
    pool_id integer REFERENCES pools (id),
    country_id integer REFERENCES countries (id),
    name text NOT NULL,
    abbreviation text,
    club text,
    seed integer,
    games_played integer,
    wins integer,
    losses integer,
    points_for integer,
    points_against integer,
    spirit_total integer,
    spirit_avg real,
    final_standing integer,
    final_standing_calculated integer
);

CREATE TABLE players (
    id integer PRIMARY KEY,
    player_id integer NOT NULL UNIQUE, -- external id
    team_id integer NOT NULL REFERENCES teams (id),
    first_name text,
    last_name text,
    jersey_num integer,
    games_played integer,
    goals integer,
    assists integer,
    callahans integer
);

CREATE TABLE games (
    id integer PRIMARY KEY,
    game_id integer NOT NULL UNIQUE, -- external id
    division_id integer REFERENCES divisions (id),
    pool_id integer REFERENCES pools (id),
    home_team_id integer REFERENCES teams (id),
    away_team_id integer REFERENCES teams (id),
    home_score integer,
    away_score integer,
    status text, -- scheduled | ongoing | completed
    scheduled_at text, -- tournament-local; no time_utc on this spec line
    field_name text,
    home_sotg integer,
    away_sotg integer
);

CREATE TABLE goals (
    id integer PRIMARY KEY,
    game_id integer NOT NULL REFERENCES games (id),
    seq integer NOT NULL, -- Goal.num
    is_home_goal integer NOT NULL, -- 0/1
    is_callahan integer NOT NULL DEFAULT 0,
    scorer_player_id integer REFERENCES players (id),
    assist_player_id integer REFERENCES players (id),
    home_score_after integer,
    away_score_after integer,
    game_time_seconds integer,
    UNIQUE (game_id, seq)
);

CREATE TABLE spirit_scores (
    id integer PRIMARY KEY,
    game_id integer NOT NULL REFERENCES games (id),
    from_team_id integer NOT NULL REFERENCES teams (id), -- team giving the score
    to_team_id integer NOT NULL REFERENCES teams (id), -- team receiving
    total integer,
    comments text,
    is_complete integer, -- TeamSpiritGame.is_complete, 0/1, nullable
    is_visible integer, -- TeamSpiritGame.is_visible, 0/1, nullable
    UNIQUE (game_id, from_team_id, to_team_id)
);

-- Keyed by category_key rather than fixed cat1..cat5 columns even though this spec line
-- always sends exactly cat1-cat5 -- keeps the read/write code identical to v1917's schema,
-- which is otherwise the same shape.
CREATE TABLE spirit_category_scores (
    spirit_score_id integer NOT NULL REFERENCES spirit_scores (id),
    category_key text NOT NULL, -- 'cat1'..'cat5'
    score integer NOT NULL,
    PRIMARY KEY (spirit_score_id, category_key)
);

