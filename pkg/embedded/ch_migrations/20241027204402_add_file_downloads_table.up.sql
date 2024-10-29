CREATE TABLE file_downloads (
  timestamp DateTime('UTC'),
  domain String,
  path String,
  visitor_id String,
  is_anon Bool ALIAS startsWith(visitor_id, 'prisme_') OR startsWith(visitor_id, 'anon_'),
  session_uuid UUID,
  session_timestamp DateTime('UTC') ALIAS UUIDv7ToDateTime(session_uuid, 'UTC'),
  session_id UInt128 ALIAS toUInt128(session_uuid),
  url String
)
ENGINE = MergeTree
ORDER BY (
  domain,
  -- Due to historical reasons, UUIDs are sorted by their second half. UUIDs
  -- should therefore not be used directly in a primary key, sorting key, or
  -- partition key of a table.
  toUInt128(session_uuid),
  timestamp,
  path,
  url,
)
PARTITION BY toUInt128(session_uuid) % 32;
