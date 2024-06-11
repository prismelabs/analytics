interval=43200 # seconds -> 12H
timestamp=$(($(date '+%s') - 7257600)) # 3 months ago
domains="'localhost', 'foo.mywebsite.localhost'"
entry_paths="'/foo', '/foo/bar', '/blog', '/blog/misc/a-nice-post'"

cat <<EOF
SELECT path, COUNT(*) AS pageviews
FROM pageviews
WHERE (timestamp >= toDateTime(${timestamp}) AND timestamp <= now())
AND session_uuid IN (
  SELECT argMax(session_uuid, pageviews) FROM sessions
  WHERE (
    (session_timestamp >= toDateTime(${timestamp}) AND session_timestamp <= now())
  OR
    (exit_timestamp >= toDateTime(${timestamp}) AND exit_timestamp <= now())
  )
  AND domain IN (${domains})
  AND entry_path IN (${entry_paths})
  GROUP BY session_uuid
)
GROUP BY path
ORDER BY pageviews DESC
EOF
