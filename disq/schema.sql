CREATE TABLE IF NOT EXISTS albums (
    id INTEGER,
    title TEXT NOT NULL,
    year INTEGER,
    rating INTEGER, -- 0 to 5
    date_added INTEGER NOT NULL, -- unix seconds
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS artists (
    id INTEGER,
    name TEXT NOT NULL,
    PRIMARY KEY (id)
);

-- M:M
CREATE TABLE IF NOT EXISTS albums_artists (
    album_id INTEGER,
    artist_id INTEGER,
    PRIMARY KEY (album_id, artist_id),
    FOREIGN KEY (album_id) REFERENCES albums (id),
    FOREIGN KEY (artist_id) REFERENCES artists (id)
);

-- album <- genre (weak?)
