RENAME TABLE sessions TO sessions_versionned;

CREATE VIEW sessions AS
SELECT
	argMax(domain, version) as domain,
	argMax(session_timestamp, version) as session_timestamp,
	argMax(entry_timestamp, version) as entry_timestamp,
	argMax(entry_path, version) as entry_path,
	argMax(exit_timestamp, version) as exit_timestamp,
	argMax(exit_path, version) as exit_path,
	argMax(visitor_id, version) as visitor_id,
	argMax(is_anon, version) as is_anon,
	argMax(session_uuid, version) as session_uuid,
	session_id,
	argMax(operating_system, version) as operating_system,
	argMax(browser_family, version) as browser_family,
	argMax(device, version) as device,
	argMax(referrer_domain, version) as referrer_domain,
	argMax(country_code, version) as country_code,
	argMax(utm_source, version) as utm_source,
	argMax(utm_medium, version) as utm_medium,
	argMax(utm_campaign, version) as utm_campaign,
	argMax(utm_term, version) as utm_term,
	argMax(utm_content, version) as utm_content,
	argMax(exit_status, version) as exit_status,
	argMax(pageviews, version) as pageviews,
	argMax(is_bounce, version) as is_bounce
FROM sessions_versionned
GROUP BY session_id;

-- Recreate pageviews materialized view as we renamed sessions to
-- sessions_versionned.
DROP TABLE pageviews_mv;
CREATE MATERIALIZED VIEW pageviews_mv TO pageviews AS
  SELECT
    exit_timestamp AS timestamp,
    domain,
    exit_path AS path,
    visitor_id,
    session_uuid,
    exit_status AS status
  FROM sessions_versionned
  WHERE sign = 1;
