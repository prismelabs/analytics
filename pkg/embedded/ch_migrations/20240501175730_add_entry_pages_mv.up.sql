CREATE TABLE entry_pages (
  timestamp DateTime('UTC'),
  domain String,
  path String,
  operating_system LowCardinality(String),
  browser_family LowCardinality(String),
  device LowCardinality(String),
  referrer_domain String,
  country_code LowCardinality(String),
  visitor_id String,
  session_id UInt64
)
ENGINE = ReplacingMergeTree(session_id)
PRIMARY KEY (domain, path, toDate(timestamp), visitor_id, session_id)
ORDER BY (domain, path, toDate(timestamp), visitor_id, session_id, operating_system, browser_family, device, referrer_domain, country_code)
SAMPLE BY session_id
PARTITION BY toYYYYMM(timestamp);

CREATE MATERIALIZED VIEW entry_pages_mv TO entry_pages AS
  SELECT
    timestamp,
    domain,
    path,
    operating_system,
    browser_family,
    device,
    referrer_domain,
    country_code,
    visitor_id,
    xxh3(timestamp, visitor_id) AS session_id
  FROM events_pageviews
  WHERE referrer_domain != domain

