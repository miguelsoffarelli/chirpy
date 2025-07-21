-- name: ResetChirps :exec
TRUNCATE TABLE chirps;

-- name: ResetUsers :exec
TRUNCATE TABLE users CASCADE;