-- Create a small materialized view containing date time of entry page for 
-- faster join in exit_pages table.
CREATE TABLE entry_pages_datetime (
  timestamp DateTime('UTC'),
  domain String,
  visitor_id String,
  session_id UInt64,
)
ENGINE = EmbeddedRocksDB(86400) -- 24h TTL
PRIMARY KEY visitor_id;

CREATE MATERIALIZED VIEW entry_pages_datetimes_mv TO entry_pages_datetime AS
  SELECT
    timestamp,
    domain,
    visitor_id,
    session_id
  FROM entry_pages;

CREATE TABLE exit_pages_no_bounce (
  timestamp DateTime('UTC'),
  domain String,
  path String,
  operating_system LowCardinality(String),
  browser_family LowCardinality(String),
  device LowCardinality(String),
  referrer_domain String,
  country_code LowCardinality(String),
  visitor_id String,
  entry_timestamp DateTime('UTC'),
  session_id UInt64
)
ENGINE = ReplacingMergeTree(session_id)
PRIMARY KEY (domain, toDate(timestamp), visitor_id, session_id)
ORDER BY (domain, toDate(timestamp), visitor_id, session_id)
SAMPLE BY session_id
PARTITION BY toYYYYMM(timestamp);

CREATE MATERIALIZED VIEW exit_pages_mv TO exit_pages_no_bounce AS
  SELECT
    events_pageviews.timestamp,
    events_pageviews.domain,
    events_pageviews.path,
    events_pageviews.operating_system,
    events_pageviews.browser_family,
    events_pageviews.device,
    events_pageviews.referrer_domain,
    events_pageviews.country_code,
    events_pageviews.visitor_id,
    entry_pages_datetime.timestamp AS entry_timestamp,
    entry_pages_datetime.session_id
  FROM events_pageviews
  INNER JOIN entry_pages_datetime
  ON entry_pages_datetime.visitor_id = events_pageviews.visitor_id
  WHERE referrer_domain == domain;

