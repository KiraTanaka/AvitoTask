SELECT true
FROM   tender t
WHERE  id = $1
    AND ( t.status = 'Published'
            OR EXISTS(SELECT 1
                        FROM   organization_responsible org_r
                            join employee emp
                                ON emp.id = org_r.user_id
                                    AND emp.username = $2
                        WHERE  org_r.organization_id = t.organization_id)
                AND t.status IN ( 'Created', 'Closed' ) )