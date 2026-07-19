package post

const (
	countQuery = `SELECT COUNT(*) FROM %s`

	listQuery = `
		SELECT id, group_id, title, url, text, raw_text, entities, invert_media, seo_title, seo_description, seo_keywords, created_at, updated_at
		FROM %s
		ORDER BY created_at DESC
		LIMIT :limit OFFSET :offset`

	allQuery = `
		SELECT id, group_id, title, url, text, raw_text, entities, invert_media, seo_title, seo_description, seo_keywords, created_at, updated_at
		FROM %s
		ORDER BY id`

	getByIDQuery = `
		SELECT id, group_id, title, url, text, raw_text, entities, invert_media, seo_title, seo_description, seo_keywords, created_at, updated_at
		FROM %s
		WHERE id = :id`

	// updateSEOQuery touches the SEO columns only, so a job writing generated
	// metadata cannot clobber a concurrent re-import of the post body.
	updateSEOQuery = `
		UPDATE %s SET
			seo_title       = :seo_title,
			seo_description = :seo_description,
			seo_keywords    = :seo_keywords,
			updated_at      = NOW()
		WHERE id = :id`

	getByURLQuery = `
		SELECT id, group_id, title, url, text, raw_text, entities, invert_media, seo_title, seo_description, seo_keywords, created_at, updated_at
		FROM %s
		WHERE url = :url`

	prevQuery = `
		SELECT id, group_id, title, url, text, raw_text, entities, invert_media, seo_title, seo_description, seo_keywords, created_at, updated_at
		FROM %s
		WHERE id < :id
		ORDER BY id DESC
		LIMIT 1`

	nextQuery = `
		SELECT id, group_id, title, url, text, raw_text, entities, invert_media, seo_title, seo_description, seo_keywords, created_at, updated_at
		FROM %s
		WHERE id > :id
		ORDER BY id ASC
		LIMIT 1`

	deleteQuery = `
		DELETE FROM %s
		WHERE id = :id`

	upsertQuery = `
		INSERT INTO %s (id, group_id, title, url, text, raw_text, entities, invert_media, seo_title, seo_description, seo_keywords, created_at, updated_at)
		VALUES (:id, :group_id, :title, :url, :text, :raw_text, CAST(:entities AS JSONB), :invert_media, :seo_title, :seo_description, :seo_keywords, :created_at, NOW())
		ON CONFLICT (id) DO UPDATE SET
			group_id = EXCLUDED.group_id,
			title = EXCLUDED.title,
			url = EXCLUDED.url,
			text = EXCLUDED.text,
			raw_text = EXCLUDED.raw_text,
			entities = EXCLUDED.entities,
			invert_media = EXCLUDED.invert_media,
			seo_title = EXCLUDED.seo_title,
			seo_description = EXCLUDED.seo_description,
			seo_keywords = EXCLUDED.seo_keywords,
			updated_at = NOW()`
)
