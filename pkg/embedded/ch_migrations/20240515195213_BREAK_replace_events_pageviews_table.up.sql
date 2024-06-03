CREATE TABLE pageviews (
  timestamp DateTime('UTC'),
  domain String,
  path String,
  visitor_id String,
  is_anon Bool ALIAS startsWith(visitor_id, 'prisme_') OR startsWith(visitor_id, 'anon_'),
  session_uuid UUID,
  session_timestamp DateTime('UTC') ALIAS UUIDv7ToDateTime(session_uuid, 'UTC'),
  session_id UInt128 ALIAS toUInt128(session_uuid),
)
ENGINE = MergeTree
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

-- Move rows to new table.
INSERT INTO pageviews
  SELECT
    timestamp,
    domain,
    path,
    visitor_id,
    toUUID('00000000-0000-7000-0000-000000000000') AS session_uuid
  FROM events_pageviews;

-- Delete old table.
DROP TABLE events_pageviews;

CREATE MATERIALIZED VIEW pageviews_mv TO pageviews AS
  SELECT
    exit_timestamp AS timestamp,
    domain,
    exit_path AS path,
    visitor_id,
    session_uuid
  FROM sessions
  WHERE sign = 1;
