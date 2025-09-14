-- name: GetMatch :one
SELECT id, player_1, player_2
FROM matches
WHERE id = $1;

-- name: CreateMatch :one
INSERT INTO matches (id, player_1, player_2, created_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetMatchScores :many
SELECT match_id, game_id, winner, played_at, winner_score, loser_score
FROM scores
WHERE match_id = $1
  AND played_at > $2
ORDER BY played_at DESC;

-- name: CreateScore :exec
INSERT INTO scores (match_id, game_id, winner, winner_score, loser_score, created_at, played_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: CreateBulkScores :copyfrom
INSERT INTO scores (match_id, game_id, winner, winner_score, loser_score, created_at, played_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetRecentScores :many
SELECT match_id, game_id, winner, played_at, winner_score, loser_score
FROM scores
WHERE match_id = $1
ORDER BY played_at DESC
LIMIT 5;
