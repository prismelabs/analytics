timestamp=$(($(date '+%s') - 7257600)) # 3 months ago
domain="'localhost', 'foo.mywebsite.localhost'"
path="'/', '/foo', '/blog'"
operating_system="'Windows', 'Linux', 'Mac OS X', 'iOS', 'Android'"
browser_family="'Firefox', 'Chrome', 'Edge', 'Opera', 'Safari'"
referrer_domain="'direct', 'twitter.com', 'facebook.com'"
country_code="'FR', 'BG', 'US'"

cat <<EOF
WITH exit_pageviews AS (
	SELECT max(timestamp) timestamp, session_id FROM events_pageviews GROUP BY session_id
)
SELECT session_id, path, COUNT(*) AS pageviews
FROM events_pageviews
WHERE timestamp >= $timestamp
  AND domain IN ($domain)
  AND path IN ($path)
  AND operating_system IN ($operating_system)
  AND browser_family IN ($browser_family)
  AND referrer_domain IN ($referrer_domain)
  AND country_code IN ($country_code)
GROUP BY path
ORDER BY pageviews DESC
EOF
