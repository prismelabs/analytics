interval=43200 # seconds -> 12H
timestamp=$(($(date '+%s') - 7257600)) # 3 months ago
domains="'localhost', 'foo.mywebsite.localhost'"
referrals="'direct', 'twitter.com', 'facebook.com'"

cat <<EOF
WITH sessions_locations AS (
  SELECT argMax(country_code, pageviews) AS code
  FROM sessions
  WHERE (
    (session_timestamp >= toDateTime(${timestamp}) AND session_timestamp <= now())
  OR
    (exit_timestamp >= toDateTime(${timestamp}) AND exit_timestamp <= now())
  )
  AND referrer_domain IN (${referrals})
  GROUP BY session_uuid
)
SELECT name AS country, COUNT(*) AS session_count
FROM sessions_locations
JOIN countries ON sessions_locations.code = countries.code
GROUP BY country
ORDER BY session_count DESC
EOF
