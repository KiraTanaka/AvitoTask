SELECT id,
	name,
	status,
	tender_id,
	author_type,
	author_id,
	version,
	created_at,
	decision
FROM bid 
WHERE id = $1