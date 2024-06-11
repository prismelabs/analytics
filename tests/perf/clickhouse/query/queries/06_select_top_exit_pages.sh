interval=43200 # seconds -> 12H
timestamp=$(($(date '+%s') - 7257600)) # 3 months ago
domains="'localhost', 'foo.mywebsite.localhost'"
entry_paths="'/foo', '/foo/bar', '/blog', '/blog/misc/a-nice-post'"
exit_paths="'/foo', '/foo/bar', '/blog', '/blog/misc/a-nice-post'"
paths="'/blog'"
operating_systems="'Windows', 'Linux', 'Mac OS X', 'iOS', 'Android'"
browsers="'Firefox', 'Chrome', 'Edge', 'Opera', 'Safari'"
referrals="'direct', 'twitter.com', 'facebook.com'"
locations="'FR', 'BG', 'US'"

cat <<EOF
WITH exit_pageviews AS (
  SELECT argMax(exit_path, pageviews) AS path
  FROM sessions
  WHERE (
    (session_timestamp >= toDateTime(${timestamp}) AND session_timestamp <= now())
  OR
    (exit_timestamp >= toDateTime(${timestamp}) AND exit_timestamp <= now())
  )
  AND domain IN (${domains})
  AND entry_path IN (${entry_paths})
  AND operating_system IN (${operating_systems})
  AND browser_family IN (${browsers})
  AND referrer_domain IN (${referrals})
  AND country_code IN (${locations})
  AND exit_path IN (${exit_paths})
  GROUP BY session_uuid
)
SELECT path, COUNT(*) AS pageviews
FROM exit_pageviews
GROUP BY path
ORDER BY pageviews DESC
EOF
