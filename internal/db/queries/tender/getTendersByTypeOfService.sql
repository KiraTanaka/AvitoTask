SELECT id,
	   name,
	   COALESCE(description,'') as description,
	   status,
	   service_type,
	   version,
	   created_at
FROM   tender
WHERE  service_type = ANY ( $1 )
		OR $2 = 0
ORDER BY name
LIMIT $3 OFFSET $4