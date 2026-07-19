-- Renames the column back. The absolute URLs are not reconstructed: the storage
-- endpoint and bucket are runtime config and unknown to SQL. Re-running the
-- media download jobs repopulates whatever the old code expected.
ALTER TABLE media RENAME COLUMN object_key TO url;
