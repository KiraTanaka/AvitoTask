SELECT TRUE
FROM employee emp
WHERE emp.username = $1
AND ('User' = $3 AND emp.id = $2
    OR 'Organization' = $3 AND EXISTS(SELECT 1
                                        FROM organization_responsible org_r
                                        WHERE org_r.organization_id = $2 AND org_r.user_id = emp.id))