SELECT params
FROM tender_version_hist
WHERE tender_id = $1 AND version = $2