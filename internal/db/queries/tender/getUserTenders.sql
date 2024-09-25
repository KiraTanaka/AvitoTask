SELECT t.id,
		t.name,
		COALESCE(t.description, '') AS description,
		t.status,
		t.service_type,
		t.version,
		t.created_at
FROM tender t
	JOIN organization_responsible org_r ON org_r.organization_id = t.organization_id
	JOIN employee e ON org_r.user_id = e.id
WHERE e.username = $1
ORDER BY name
		LIMIT $2 OFFSET $3