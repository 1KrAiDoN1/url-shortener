CREATE TABLE IF NOT EXISTS url (
	id SERIAL PRIMARY KEY,
	alias TEXT NOT NULL UNIQUE,
	url TEXT NOT NULL,
    UNIQUE(url)
);

CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);