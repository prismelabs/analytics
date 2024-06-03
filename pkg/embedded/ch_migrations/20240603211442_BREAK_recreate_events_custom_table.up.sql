RENAME TABLE events_custom TO events_custom_old;

-- Add visitor_id to primary key so it can be used for sampling.

-- Create a more optimized version of events_custom table.
-- This version use a better order by clause and adds a partition by clause
-- that can't be added to an existing table.

-- It is beneficial for queries to order the primary key columns by cardinality in ascending order.
-- https://clickhouse.com/docs/en/optimize/sparse-primary-indexes#summary
CREATE TABLE events_custom (
  timestamp DateTime('UTC'),
  domain String,
  path String,
  visitor_id String,
  is_anon Bool ALIAS startsWith(visitor_id, 'prisme_') OR startsWith(visitor_id, 'anon_'),
  session_uuid UUID,
  session_timestamp DateTime('UTC') ALIAS UUIDv7ToDateTime(session_uuid, 'UTC'),
  session_id UInt128 ALIAS toUInt128(session_uuid),
  -- event name
  name String,
  -- JSON keys and values string
  keys Array(String),
  values Array(String)
)
ENGINE = MergeTree
ORDER BY (
  domain,
  -- Due to historical reasons, UUIDs are sorted by their second half. UUIDs
  -- should therefore not be used directly in a primary key, sorting key, or
  -- partition key of a table.
  toUInt128(session_uuid),
  name,
  timestamp,
  path
)
PARTITION BY toUInt128(session_uuid) % 32;

-- Move rows to new table.
INSERT INTO events_custom
  SELECT
    timestamp,
    domain,
    path,
    visitor_id,
    toUUID('00000000-0000-7000-0000-000000000000') AS session_uuid,
    name,
    keys,
    values
  FROM events_custom_old;

-- Delete old table.
DROP TABLE events_custom_old;
