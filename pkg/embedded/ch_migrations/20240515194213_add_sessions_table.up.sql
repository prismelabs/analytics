CREATE TABLE sessions (
  domain String,
  session_timestamp DateTime('UTC') ALIAS UUIDv7ToDateTime(session_uuid, 'UTC'),
  entry_timestamp DateTime('UTC') ALIAS session_timestamp,
  entry_path String,
  exit_timestamp DateTime('UTC'),
  exit_path String,
  visitor_id String,
  is_anon Bool ALIAS startsWith(visitor_id, 'prisme_') OR startsWith(visitor_id, 'anon_'),
  session_uuid UUID,
  session_id UInt128 ALIAS toUInt128(session_uuid),
  operating_system LowCardinality(String),
  browser_family LowCardinality(String),
  device LowCardinality(String),
  referrer_domain String,
  country_code LowCardinality(String),
  utm_source String,
  utm_medium String,
  utm_campaign String,
  utm_term String,
  utm_content String,
  version UInt16,
  pageviews UInt16 ALIAS version,
  is_bounce Bool ALIAS pageviews = 1,
  sign Int8
)
ENGINE = VersionedCollapsingMergeTree(sign, version)
ORDER BY (
  domain,
  entry_path,
  operating_system,
  browser_family,
  device,
  referrer_domain,
  country_code,
  utm_source,
  utm_medium,
  utm_campaign,
  utm_term,
  utm_content,
  -- Due to historical reasons, UUIDs are sorted by their second half. UUIDs
  -- should therefore not be used directly in a primary key, sorting key, or
  -- partition key of a table.
  toUInt128(session_uuid),
  -- Exit pageviews data isn't part of sorting key as only rows
  -- sharing the same sorting keys are collapsed.
)
PARTITION BY toYYYYMM(exit_timestamp);
