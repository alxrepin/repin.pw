ALTER TABLE media RENAME COLUMN url TO object_key;

UPDATE media
SET object_key = regexp_replace(object_key, '^https?://[^/]+/[^/]+/', '')
WHERE object_key ~ '^https?://';

UPDATE channels
SET avatar = regexp_replace(avatar, '^https?://[^/]+/[^/]+/', '')
WHERE avatar ~ '^https?://';
