-- name: CreateAuthor :one
INSERT INTO authors (name) VALUES (?)
RETURNING *;

-- name: GetAuthor :one
SELECT * FROM authors WHERE id = ?;

-- name: ListAuthors :many
SELECT * FROM authors ORDER BY id;

-- name: CountAuthors :one
SELECT COUNT(*) FROM authors;

-- name: CreateBook :one
INSERT INTO books (title, author_id) VALUES (?, ?)
RETURNING *;

-- name: ListBooksByAuthor :many
SELECT * FROM books WHERE author_id = ? ORDER BY id;
