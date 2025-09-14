CREATE TABLE scores (
  match_id uuid NOT NULL
    REFERENCES matches (id),
  game_id uuid NOT NULL,
  winner TEXT NOT NULL,
  score TEXT NOT NULL,
  created_at timestamptz NOT NULL,
  PRIMARY KEY(match_id, game_id)
);
