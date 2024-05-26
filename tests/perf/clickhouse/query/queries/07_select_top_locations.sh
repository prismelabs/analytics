timestamp=$(($(date '+%s') - 7257600)) # 3 months ago
domain="'localhost', 'foo.mywebsite.localhost'"
path="'/', '/foo', '/blog'"
operating_system="'Windows', 'Linux', 'Mac OS X', 'iOS', 'Android'"
browser_family="'Firefox', 'Chrome', 'Edge', 'Opera', 'Safari'"
referrer_domain="'direct', 'twitter.com', 'facebook.com'"
country_code="'FR', 'BG', 'US'"

cat <<EOF
SELECT DISTINCT(name) AS country, COUNT(*) AS pageview
FROM entry_pageviews
JOIN countries ON entry_pageviews.country_code = countries.code
WHERE timestamp >= $timestamp
  AND domain IN ($domain)
  AND path IN ($path)
  AND operating_system IN ($operating_system)
  AND browser_family IN ($browser_family)
  AND referrer_domain IN ($referrer_domain)
  AND country_code IN ($country_code)
GROUP BY country
ORDER BY pageview DESC
EOF
