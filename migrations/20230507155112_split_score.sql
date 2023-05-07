-- Add migration script here
ALTER TABLE scores DROP COLUMN score;

ALTER TABLE scores ADD winner_score SMALLINT;
ALTER TABLE scores ADD loser_score SMALLINT;
