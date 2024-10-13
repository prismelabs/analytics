RENAME TABLE pageviews TO pageviews_old;

CREATE TABLE pageviews (
  timestamp DateTime('UTC'),
  domain String,
  path String,
  visitor_id String,
  time_on_page UInt16,
  is_anon Bool ALIAS startsWith(visitor_id, 'prisme_') OR startsWith(visitor_id, 'anon_'),
  session_uuid UUID,
  session_timestamp DateTime('UTC') ALIAS UUIDv7ToDateTime(session_uuid, 'UTC'),
  session_id UInt128 ALIAS toUInt128(session_uuid),
  version UInt16,
  sign Int8,
)
ENGINE = VersionedCollapsingMergeTree(sign, version)
ORDER BY (
  domain,
  -- Due to historical reasons, UUIDs are sorted by their second half. UUIDs
  -- should therefore not be used directly in a primary key, sorting key, or
  -- partition key of a table.
  toUInt128(session_uuid),
  timestamp,
  path
)
PARTITION BY toUInt128(session_uuid) % 32;

ALTER TABLE sessions ADD COLUMN time_on_exit_page_seconds UInt16;
ALTER TABLE sessions ADD COLUMN total_time_on_page_seconds UInt16;

-- Replace alias with real columns.
ALTER TABLE sessions DROP COLUMN is_bounce; -- depends on pageviews
ALTER TABLE sessions DROP COLUMN pageview_count;
ALTER TABLE sessions DROP COLUMN pageviews;
ALTER TABLE sessions ADD COLUMN pageview_count UInt16;
ALTER TABLE sessions ADD COLUMN pageviews UInt16 ALIAS pageview_count;
ALTER TABLE sessions ADD COLUMN is_bounce Bool ALIAS pageview_count = 1;

INSERT INTO pageviews
  SELECT
    timestamp,
    domain,
    path,
    visitor_id,
    0 AS time_on_page,
    session_uuid,
    1 as version,
    1 as sign
  FROM pageviews_old;

DROP TABLE pageviews_mv;
DROP TABLE pageviews_old;

CREATE MATERIALIZED VIEW pageviews_mv TO pageviews AS
  SELECT
    exit_timestamp AS timestamp,
    domain,
    exit_path AS path,
    visitor_id,
    time_on_exit_page_seconds AS time_on_page_seconds,
    session_uuid,
    version,
    sign
  FROM sessions
  WHERE sign = 1;
