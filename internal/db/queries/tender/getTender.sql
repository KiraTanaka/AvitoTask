SELECT id,
	name,
	COALESCE(description,'') as description,
	status,
	service_type,
	organization_id,
	version,
	created_at
FROM tender 
WHERE id = $1