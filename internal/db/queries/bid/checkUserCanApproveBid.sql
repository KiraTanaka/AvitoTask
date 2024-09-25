SELECT TRUE
FROM tender t
    JOIN organization_responsible org_r ON org_r.organization_id = t.organization_id
    JOIN employee emp ON emp.id = org_r.user_id AND emp.username = $2
WHERE t.id = $1