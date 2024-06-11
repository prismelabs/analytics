interval=43200 # seconds -> 12H
timestamp=$(($(date '+%s') - 7257600)) # 3 months ago
domains="'localhost', 'foo.mywebsite.localhost'"
browsers="'Firefox', 'Chrome', 'Edge', 'Opera', 'Safari'"

cat <<EOF
WITH entry_pageviews AS (
  SELECT argMax(entry_path, pageviews) AS path
  FROM sessions
  WHERE (
    (session_timestamp >= toDateTime(${timestamp}) AND session_timestamp <= now())
  OR
    (exit_timestamp >= toDateTime(${timestamp}) AND exit_timestamp <= now())
  )
  AND domain IN (${domains})
  AND browser_family IN (${browsers})
  GROUP BY session_uuid
)
SELECT path, COUNT(*) AS session_count
FROM entry_pageviews
GROUP BY path
ORDER BY session_count DESC
EOF
