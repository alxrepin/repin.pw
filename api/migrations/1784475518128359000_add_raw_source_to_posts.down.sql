BEGIN;

ALTER TABLE posts
    DROP COLUMN raw_text,
    DROP COLUMN entities;

COMMIT;
