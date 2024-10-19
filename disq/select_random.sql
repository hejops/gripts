SELECT
    rand.title AS album,
    -- Usage notes: concat() and concat_ws() are appropriate for concatenating
    -- the values of multiple columns within the same row, while group_concat()
    -- joins together values from different rows.
    -- https://impala.apache.org/docs/build/asf-site-html/topics/impala_group_concat.html
    group_concat(artists.name, ' ') AS artist
FROM
    (
        -- TODO: qualifying the refs (e.g. albums.id) breaks random func for no
        -- reason
        -- https://docs.sqlfluff.com/en/stable/reference/rules.html#column-references-should-be-qualified-consistently-in-single-table-statements
        SELECT
            albums.id,
            albums.title
        FROM albums
        WHERE albums.rating >= 3
        -- AND id = 2691647
        ORDER BY random() LIMIT 1
    ) AS rand
INNER JOIN albums_artists
    ON rand.id = albums_artists.album_id
INNER JOIN artists
    ON albums_artists.artist_id = artists.id
GROUP BY album
