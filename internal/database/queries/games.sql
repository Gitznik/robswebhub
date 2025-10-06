-- name: CreateGame :one
insert into games (game_id, name, type, created_at, updated_at)
values($1, $2, $3, $4, $4)
returning *;

-- name: CreateGameMembership :copyfrom
insert into games_membership (game_id, player_id, role, created_at, updated_at)
values ($1, $2, $3, $4, $4);

-- name: CreatePlayer :one
insert into players (player_id, created_at, updated_at)
values ($1, $2, $2)
returning *;

-- name: GetPlayerInformation :one
select player_id
from players
where player_id = $1;

-- name: ListGamesOfUser :many
select games.game_id, games.name, games.type, gm.role
from games_membership gm
left join games
on gm.game_id = games.game_id
where gm.game_id = games.game_id
  and gm.player_id = $1;
