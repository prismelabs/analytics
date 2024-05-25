-- Sessions / visits table. Sessions are implicitly created when a pageview
-- event with an external or direct referrer domain is sent.
CREATE TABLE sessions (
  timestamp DateTime('UTC'),
  domain String,
  path String,
  operating_system LowCardinality(String),
  browser_family LowCardinality(String),
  device LowCardinality(String),
  referrer_domain String,
  country_code LowCardinality(String),
  visitor_id String,
  session_uuid UUID, -- UUIDv7
  -- Due to historical reasons, UUIDs are sorted by their second half. UUIDs
  -- should therefore not be used directly in a primary key, sorting key, or
  -- partition key of a table.
  -- This alias is a workaround of these limitations.
  session_id UInt128 ALIAS toUInt128(session_uuid)
)
ENGINE = MergeTree
PRIMARY KEY (domain, path, toDate(timestamp))
ORDER BY (
  domain,
  path,
  toDate(timestamp),
  visitor_id,
  referrer_domain,
  operating_system,
  browser_family,
  device,
  country_code
)
PARTITION BY toYYYYMM(timestamp);

-- Populate table.
-- generateUUIDv4 as v7 isn't available in current stable release but prisme
-- will insert UUIDs v7.
INSERT INTO sessions
  SELECT timestamp, domain, path, operating_system, browser_family, device,
    referrer_domain, country_code, visitor_id, generateUUIDv4() as session_uuid
  FROM pageviews
  WHERE referrer_domain != domain;

-- Recreate pageviews table.
-- DROP is fine as this table has a NULL engine. It's pageviews table that
-- contains the actual data.
DROP TABLE events_pageviews;
DROP TABLE pageviews_mv;
CREATE TABLE events_pageviews(
  timestamp DateTime('UTC'),
  domain String,
  path String,
  visitor_id String,
  session_id UInt128
)
ENGINE = MergeTree
ORDER BY (session_id, timestamp)
PARTITION BY toYYYYMM(timestamp);

-- Drop materialized views based on events_pageviews.
DROP TABLE entry_pageviews_mv;
DROP TABLE exit_pageviews_no_bounce_mv;

-- Drop entry_pageviews as sessions contains same information.
DROP TABLE entry_pageviews;

-- Not used anymore.
DROP TABLE exit_pageviews_no_bounce;

-- Move rows to new table.
INSERT INTO events_pageviews
  SELECT timestamp, domain, path, visitor_id, 0 AS session_id
  FROM pageviews;

DROP TABLE pageviews;

-- Insert entry pages into pageviews table.
CREATE MATERIALIZED VIEW events_pageviews_mv TO events_pageviews AS
  SELECT timestamp, domain, path, visitor_id, session_id
  FROM prisme.sessions;

-- Entry pages views.
CREATE VIEW prisme.entry_pageviews AS
  SELECT timestamp, domain, path, visitor_id, session_id
  FROM prisme.sessions;

-- Recreate custom events table.
RENAME TABLE events_custom TO events_custom_old;

CREATE TABLE events_custom (
  timestamp DateTime('UTC'),
  domain String,
  path String,
  visitor_id String,
  session_id UInt128,
  -- event name
  name String,
  -- JSON keys and values string
  keys Array(String),
  values Array(String)
)
ENGINE = MergeTree
ORDER BY (name, session_id, timestamp)
PARTITION BY toYYYYMM(timestamp);

-- Move rows to new table.
INSERT INTO events_custom
  SELECT
    timestamp,
    domain,
    path,
    visitor_id,
    session_id,
    name,
    keys,
    values
  FROM events_custom_old;

DROP TABLE events_custom_old;
