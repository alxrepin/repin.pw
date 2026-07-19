BEGIN;

ALTER TABLE posts
    ADD COLUMN raw_text TEXT,
    ADD COLUMN entities JSONB;

COMMIT;
