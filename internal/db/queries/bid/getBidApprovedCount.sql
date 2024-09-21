SELECT COUNT(*)
FROM bid_decision
WHERE bid_id = $1 AND decision = 'Approved'