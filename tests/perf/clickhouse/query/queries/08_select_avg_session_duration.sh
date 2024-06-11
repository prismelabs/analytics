interval=43200 # seconds -> 12H
timestamp=$(($(date '+%s') - 7257600)) # 3 months ago
domains="'localhost', 'foo.mywebsite.localhost'"
entry_paths="'/foo', '/foo/bar', '/blog', '/blog/misc/a-nice-post'"

cat <<EOF
WITH sessions_duration AS (
  SELECT argMax(exit_timestamp, pageviews) - argMax(session_timestamp, pageviews) AS duration
  FROM sessions
  WHERE (
    (session_timestamp >= toDateTime(${timestamp}) AND session_timestamp <= now())
  OR
    (exit_timestamp >= toDateTime(${timestamp}) AND exit_timestamp <= now())
  )
  AND domain IN (${domains})
  AND entry_path IN (${entry_paths})
  AND session_timestamp != exit_timestamp
  GROUP BY session_uuid
)
SELECT avg(duration) AS "Average session duration"
FROM sessions_duration
EOF
