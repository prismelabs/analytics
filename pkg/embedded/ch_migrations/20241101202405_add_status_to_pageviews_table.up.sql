ALTER TABLE sessions ADD COLUMN exit_status UInt16 DEFAULT 200 AFTER utm_content;
ALTER TABLE pageviews ADD COLUMN status UInt16 DEFAULT 200 AFTER session_uuid;

DROP TABLE pageviews_mv;

CREATE MATERIALIZED VIEW pageviews_mv TO pageviews AS
  SELECT
    exit_timestamp AS timestamp,
    domain,
    exit_path AS path,
    visitor_id,
    session_uuid,
    exit_status AS status
  FROM sessions
  WHERE sign = 1;

