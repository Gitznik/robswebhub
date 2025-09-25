CREATE TABLE players (
  player_id text not null,
  game_limit smallint not null default 10,
  scores_limit smallint not null default 1000,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  PRIMARY KEY(player_id)
);

CREATE TYPE "game_type" AS ENUM (
  'durak'
);

CREATE TYPE "game_role" AS ENUM (
  'owner',
  'writer',
  'reader'
);

CREATE TABLE games (
  game_id uuid NOT NULL,
  name text not null,
  type game_type not null,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  PRIMARY KEY(game_id)
);

CREATE TABLE games_membership (
  game_id uuid NOT NULL
    REFERENCES games (game_id),
  player_id text not null
    REFERENCES players (player_id),
  role game_role not null,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  PRIMARY KEY(game_id, player_id)
);
