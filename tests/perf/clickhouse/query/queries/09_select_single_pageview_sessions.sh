interval=43200 # seconds -> 12H
timestamp=$(($(date '+%s') - 7257600)) # 3 months ago
domains="'localhost', 'foo.mywebsite.localhost'"
operating_systems="'Windows', 'Linux', 'Mac OS X', 'iOS', 'Android'"

cat <<EOF
WITH bounces AS (
  SELECT argMax(pageviews, pageviews) AS pageviews
  FROM sessions
  WHERE (
    (session_timestamp >= toDateTime(${timestamp}) AND session_timestamp <= now())
  OR
    (exit_timestamp >= toDateTime(${timestamp}) AND exit_timestamp <= now())
  )
  AND domain IN (${domains})
  AND operating_system IN (${operating_systems})
  GROUP BY session_uuid
  HAVING pageviews = 1
)
SELECT COUNT(*) AS bounces FROM bounces
EOF
