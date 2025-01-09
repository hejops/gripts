-- artists with more than 3 ratings, and with average rating 2.7 (out of 5)

-- assign query result to table
-- https://www.postgresql.org/docs/current/queries-with.html#QUERIES-WITH-SELECT
WITH joined AS (
    SELECT *
    FROM albums
    INNER JOIN albums_artists ON albums.id = albums_artists.album_id
    INNER JOIN artists ON albums_artists.artist_id = artists.id
    -- WHERE albums.rating >= 3
    -- note: not sorting makes the query at least 25x slower!
    ORDER BY albums.rating DESC
),

at_least_three_ratings AS (
    SELECT
        name,
        title,
        rating
    FROM joined
    WHERE
        (
            -- 'group'
            -- https://www.machinelearningplus.com/sql/how-to-get-top-n-results-in-each-group-by-group-in-sql/
            SELECT count(*) FROM joined AS j
            WHERE joined.name = j.name
        )
        >= 3
    ORDER BY joined.name ASC, joined.rating DESC
)

SELECT
    name,
    avg(rating) AS rs
FROM at_least_three_ratings
GROUP BY name
HAVING avg(rating) >= 2.7 -- jsb ~ 2.77
