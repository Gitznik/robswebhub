CREATE TABLE scores (
  match_id uuid NOT NULL
    REFERENCES matches (id),
  game_id uuid NOT NULL,
  winner TEXT NOT NULL,
  winner_score smallint NOT NULL,
  loser_score smallint NOT NULL,
  created_at timestamptz NOT NULL,
  played_at DATE NOT NULL,
  PRIMARY KEY(match_id, game_id)
);
