interval=43200 # seconds -> 12H
timestamp=$(($(date '+%s') - 7257600)) # 3 months ago
domains="'localhost', 'foo.mywebsite.localhost'"
locations="'FR', 'BG', 'US'"
paths="'/foo', '/foo/bar', '/blog', '/blog/misc/a-nice-post'"

cat <<EOF
SELECT toStartOfInterval(timestamp, INTERVAL ${interval} second) AS time, COUNT(*)
FROM pageviews
WHERE timestamp >= toDateTime(${timestamp})
AND timestamp <= now()
AND domain IN (${domains})
AND path IN (${paths})
AND session_uuid IN (
  SELECT session_uuid FROM sessions
  WHERE (
    (session_timestamp >= toDateTime(${timestamp}) AND session_timestamp <= now())
  OR
    (exit_timestamp >= toDateTime(${timestamp}) AND exit_timestamp <= now())
  )
  AND domain IN (${domains})
  AND country_code IN (${locations})
  GROUP BY session_uuid
)
GROUP BY time
ORDER BY time
EOF
