INSERT INTO tender
			(name,
			description,
			service_type,
			status,
			organization_id,
			version,
			created_at)
VALUES     ($1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7)
RETURNING id