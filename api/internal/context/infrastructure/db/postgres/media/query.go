package media

const (
	getByMessageIDQuery = `
		SELECT tg_message_id, post_id, file_id, type, object_key, mime_type, file_name, size_bytes, width, height, duration, created_at, updated_at
		FROM %s
		WHERE tg_message_id = :tg_message_id`

	listByPostIDsQuery = `
		SELECT tg_message_id, post_id, file_id, type, object_key, mime_type, file_name, size_bytes, width, height, duration, created_at, updated_at
		FROM %s
		WHERE post_id IN (?)
		ORDER BY post_id, tg_message_id`

	upsertQuery = `
		INSERT INTO %s (tg_message_id, post_id, file_id, type, object_key, mime_type, file_name, size_bytes, width, height, duration, created_at, updated_at)
		VALUES (:tg_message_id, :post_id, :file_id, :type, :object_key, :mime_type, :file_name, :size_bytes, :width, :height, :duration, :created_at, NOW())
		ON CONFLICT (tg_message_id) DO UPDATE SET
			post_id = EXCLUDED.post_id,
			file_id = EXCLUDED.file_id,
			type = EXCLUDED.type,
			object_key = EXCLUDED.object_key,
			mime_type = EXCLUDED.mime_type,
			file_name = EXCLUDED.file_name,
			size_bytes = EXCLUDED.size_bytes,
			width = EXCLUDED.width,
			height = EXCLUDED.height,
			duration = EXCLUDED.duration,
			updated_at = NOW()`

	deleteStaleQuery = `
		DELETE FROM %s
		WHERE post_id = ? AND tg_message_id NOT IN (?)`

	deleteByPostIDQuery = `
		DELETE FROM %s
		WHERE post_id = :post_id`

	deleteByMessageIDQuery = `
		DELETE FROM %s
		WHERE tg_message_id = :tg_message_id`
)
