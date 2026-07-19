-- Media used to store an absolute URL built from the storage endpoint, which
-- pinned rows to a particular deployment and leaked the internal host to
-- clients. Keep only the bucket key; the API builds the public link.
ALTER TABLE media RENAME COLUMN url TO object_key;

-- Strip "scheme://host/bucket/" from existing rows, leaving "media/123.jpg".
UPDATE media
SET object_key = regexp_replace(object_key, '^https?://[^/]+/[^/]+/', '')
WHERE object_key ~ '^https?://';

UPDATE channels
SET avatar = regexp_replace(avatar, '^https?://[^/]+/[^/]+/', '')
WHERE avatar ~ '^https?://';
