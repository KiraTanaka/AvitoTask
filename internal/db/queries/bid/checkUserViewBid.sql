SELECT true
FROM bid b
WHERE b.id = $1
AND (b.status IN ('Created', 'Canceled')
        AND (author_type = 'User' AND EXISTS(SELECT 1
                                            FROM employee emp
                                            WHERE emp.id = b.author_id AND emp.username = $2)
                OR b.author_type = 'Organization'
                    AND EXISTS(SELECT 1
                                FROM organization_responsible org_r
                                    JOIN employee emp ON emp.id = org_r.user_id AND emp.username = $2
                                WHERE org_r.organization_id = b.author_id))
    OR
    b.status = 'Published')