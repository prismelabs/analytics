interval=43200 # seconds -> 12H
timestamp=$(($(date '+%s') - 7257600)) # 3 months ago
domains="'localhost', 'foo.mywebsite.localhost'"
referrals="'direct', 'twitter.com', 'facebook.com'"

cat <<EOF
SELECT COUNT(DISTINCT(visitor_id)) AS "Live visitors"
FROM sessions
WHERE addMinutes(exit_timestamp, 15) > now()
AND domain IN (${domains})
AND referrer_domain IN (${referrals})
EOF
