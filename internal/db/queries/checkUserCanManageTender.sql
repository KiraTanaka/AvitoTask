SELECT true
FROM organization_responsible org_r
	JOIN employee emp ON emp.id = org_r.user_id
WHERE org_r.organization_id = $1 AND emp.username = $2