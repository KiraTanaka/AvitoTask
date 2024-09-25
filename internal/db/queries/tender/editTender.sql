UPDATE tender
SET    name = :name,
		description = :description,
		service_type = :service_type
WHERE  id = :id