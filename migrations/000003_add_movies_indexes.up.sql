CREATE INDEX IF NOT EXISTS movies_title_idx ON movies USING GIN (to_tsquery('simple', title));
CREATE INDEX IF NOT EXISTS movies_genres_idx ON movies USING GIN (genres);