INSERT INTO bid_decision
			(bid_id,
			username,
			decision)
VALUES     ($1,
			$2,
			$3)
RETURNING id