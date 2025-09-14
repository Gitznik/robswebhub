ALTER TABLE scores DROP COLUMN winner_score;
ALTER TABLE scores DROP COLUMN loser_score;
ALTER TABLE scores ADD score TEXT;
