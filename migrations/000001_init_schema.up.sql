CREATE TABLE IF NOT EXISTS users (
    id       SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email    TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS pages (
    title        TEXT PRIMARY KEY,
    url          TEXT NOT NULL UNIQUE,
    language     TEXT NOT NULL CHECK (language IN ('en', 'da')) DEFAULT 'en',
    last_updated TIMESTAMP,
    content      TEXT NOT NULL,
    search_vector TSVECTOR
);

CREATE INDEX IF NOT EXISTS pages_search_idx ON pages USING GIN (search_vector);

CREATE OR REPLACE FUNCTION pages_tsvector_update() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector := to_tsvector('english', COALESCE(NEW.title, '') || ' ' || COALESCE(NEW.content, ''));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER pages_tsvector_trigger
    BEFORE INSERT OR UPDATE ON pages
    FOR EACH ROW EXECUTE FUNCTION pages_tsvector_update();
