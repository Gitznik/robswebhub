CREATE TABLE matches(
  id uuid NOT NULL,
  PRIMARY KEY (id),
  player_1 TEXT NOT NULL,
  player_2 TEXT NOT NULL,
  created_at timestamptz NOT NULL
);
