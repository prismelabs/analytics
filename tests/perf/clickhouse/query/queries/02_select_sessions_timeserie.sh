interval=43200 # seconds -> 12H
timestamp=$(($(date '+%s') - 7257600)) # 3 months ago
domains="'localhost', 'foo.mywebsite.localhost'"
locations="'FR', 'BG', 'US'"

cat <<EOF
WITH exit_timestamps AS (
  SELECT
    argMax(exit_timestamp, pageviews) AS timestamp,
    argMax(visitor_id, pageviews) AS visitor_id
  FROM sessions
  WHERE (
    (session_timestamp >= toDateTime(${timestamp}) AND session_timestamp <= now())
  OR
    (exit_timestamp >= toDateTime(${timestamp}) AND exit_timestamp <= now())
  )
  AND domain IN (${domains})
  AND country_code IN (${locations})
  GROUP BY session_uuid
)
SELECT toStartOfInterval(timestamp, INTERVAL ${interval} second) AS time, COUNT(*)
FROM exit_timestamps
GROUP BY time
ORDER BY time
EOF
