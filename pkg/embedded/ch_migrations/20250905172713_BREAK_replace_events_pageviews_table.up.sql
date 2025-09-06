RENAME TABLE pageviews TO pageviews_old;

CREATE TABLE pageviews (
  timestamp DateTime('UTC'),
  domain String,
  path String,
  visitor_id String,
  is_anon Bool ALIAS startsWith(visitor_id, 'prisme_') OR startsWith(visitor_id, 'anon_'),
  session_uuid UUID,
  status UInt16,
  session_timestamp DateTime('UTC') ALIAS UUIDv7ToDateTime(session_uuid, 'UTC'),
  session_id UInt128 ALIAS toUInt128(session_uuid),
)
ENGINE = MergeTree
ORDER BY (
  domain,
  path,
  toDate(timestamp),
  -- Due to historical reasons, UUIDs are sorted by their second half. UUIDs
  -- should therefore not be used directly in a primary key, sorting key, or
  -- partition key of a table.
  toUInt128(session_uuid)
)
PARTITION BY toYYYYMM(UUIDv7ToDateTime(session_uuid, 'UTC'));

-- Move rows to new table.
INSERT INTO pageviews
  SELECT
    timestamp,
    domain,
    path,
    visitor_id,
    session_uuid,
    status
  FROM pageviews_old;

-- Delete old table.
DROP TABLE pageviews_old;
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
