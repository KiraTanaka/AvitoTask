SELECT b.id,
	b.name,
	b.status,
	b.author_type,
	b.author_id,
	b.version,
	b.created_at
FROM bid b
WHERE (author_type = 'User' AND exists(select 1
								from employee emp
								where emp.id = b.author_id and emp.username= $1)
	OR b.author_type = 'Organization'
		AND EXISTS(SELECT 1
					FROM organization_responsible org_r
						JOIN employee emp ON emp.id = org_r.user_id AND emp.username = $1
					WHERE org_r.organization_id = b.author_id))
ORDER BY name
LIMIT $2 OFFSET $3