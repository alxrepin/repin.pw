BEGIN;

ALTER TABLE posts
    DROP COLUMN invert_media;

COMMIT;
