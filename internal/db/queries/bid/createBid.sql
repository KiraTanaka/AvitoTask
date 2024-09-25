INSERT INTO bid
			(name,
			description,
			status,
			tender_id,
			author_type,
			author_id,
			version,
			created_at)
VALUES     ($1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8)
		RETURNING id