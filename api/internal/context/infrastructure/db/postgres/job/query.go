package job

const (
	enqueueQuery = `
		INSERT INTO %s (kind, dedup_key, payload, max_attempts, run_at)
		VALUES (:kind, :dedup_key, CAST(:payload AS JSONB), :max_attempts, NOW())
		ON CONFLICT (dedup_key) WHERE status = 'pending' DO UPDATE SET
			payload      = EXCLUDED.payload,
			max_attempts = EXCLUDED.max_attempts,
			run_at       = LEAST(jobs.run_at, EXCLUDED.run_at),
			attempts     = 0,
			last_error   = NULL,
			updated_at   = NOW()`

	claimQuery = `
		UPDATE %s SET
			status     = 'running',
			attempts   = attempts + 1,
			locked_at  = NOW(),
			updated_at = NOW()
		WHERE id = (
			SELECT id FROM %s
			WHERE status = 'pending'
			  AND run_at <= NOW()
			  AND kind = ANY(:kinds)
			ORDER BY run_at, id
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, kind, dedup_key, payload, status, attempts, max_attempts, run_at, last_error, created_at`

	completeQuery = `
		DELETE FROM %s
		WHERE id = :id`

	retryQuery = `
		UPDATE %s SET
			status     = 'pending',
			locked_at  = NULL,
			run_at     = :run_at,
			last_error = :last_error,
			updated_at = NOW()
		WHERE id = :id
		  AND NOT EXISTS (
			  SELECT 1 FROM %s newer
			  WHERE newer.dedup_key = jobs.dedup_key
				AND newer.status = 'pending'
				AND newer.id <> jobs.id
		  )`

	dropSupersededQuery = `
		DELETE FROM %s
		WHERE id = :id AND status = 'running'`

	buryQuery = `
		UPDATE %s SET
			status     = 'failed',
			locked_at  = NULL,
			last_error = :last_error,
			updated_at = NOW()
		WHERE id = :id`

	requeueStaleQuery = `
		UPDATE %s SET
			status     = 'pending',
			locked_at  = NULL,
			run_at     = NOW(),
			last_error = 'reclaimed: worker lease expired',
			updated_at = NOW()
		WHERE id IN (
			SELECT DISTINCT ON (stale.dedup_key) stale.id
			FROM %s stale
			WHERE stale.status = 'running'
			  AND stale.locked_at < :threshold
			  AND NOT EXISTS (
				  SELECT 1 FROM %s newer
				  WHERE newer.dedup_key = stale.dedup_key
					AND newer.status = 'pending'
			  )
			ORDER BY stale.dedup_key, stale.id
		)`

	dropStaleSupersededQuery = `
		DELETE FROM %s
		WHERE status = 'running' AND locked_at < :threshold`
)
