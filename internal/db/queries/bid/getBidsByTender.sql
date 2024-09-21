SELECT id,
	name,
	status,
	author_type,
	author_id,
	version,
	created_at
FROM   bid
WHERE tender_id = $1
ORDER BY name
LIMIT $2 OFFSET $3