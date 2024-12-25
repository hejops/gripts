SELECT title
FROM artists
INNER JOIN albums_artists
    ON artists.id = albums_artists.artist_id
INNER JOIN albums
    ON albums_artists.album_id = albums.id
-- ? will be substituted in go; https://go.dev/doc/database/sql-injection
WHERE name = ?
ORDER BY random() LIMIT 1
