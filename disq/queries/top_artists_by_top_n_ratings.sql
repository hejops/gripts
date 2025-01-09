-- artists whose top 3 ratings (out of 5) add up to 11 or more (out of 15)

WITH

joined AS (
    SELECT *
    FROM albums
    INNER JOIN albums_artists ON albums.id = albums_artists.album_id
    INNER JOIN artists ON albums_artists.artist_id = artists.id
),


-- https://www.machinelearningplus.com/sql/how-to-get-top-n-results-in-each-group-by-group-in-sql/
top_three AS (
    SELECT
        name,
        title,
        rating,
        row_number() OVER (
            PARTITION BY name
            ORDER BY rating DESC
        ) AS rn
    FROM joined
    -- WHERE rn <= 3 -- not available in this scope
)

SELECT
    name,
    sum(rating) AS sum
FROM top_three
WHERE rn <= 3
GROUP BY name
HAVING sum(rating) >= 11
ORDER BY sum DESC, name ASC
