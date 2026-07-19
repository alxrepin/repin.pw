package channel

const (
	getQuery = `
		SELECT id, name, title, description, avatar, subscriptions, last_message_id, created_at
		FROM %s
		ORDER BY id
		LIMIT 1`

	upsertQuery = `
		INSERT INTO %s (id, name, title, description, avatar, subscriptions, last_message_id, created_at)
		VALUES (:id, :name, :title, :description, :avatar, :subscriptions, :last_message_id, :created_at)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			avatar = EXCLUDED.avatar,
			subscriptions = EXCLUDED.subscriptions,
			last_message_id = GREATEST(%s.last_message_id, EXCLUDED.last_message_id)`
)
