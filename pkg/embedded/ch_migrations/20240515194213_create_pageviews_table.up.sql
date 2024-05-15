RENAME TABLE events_pageviews TO pageviews;

CREATE TABLE events_pageviews (
	timestamp DateTime('UTC'),
	domain String,
	path String,
	operating_system LowCardinality(String),
	browser_family LowCardinality(String),
	device LowCardinality(String),
	referrer_domain String,
	country_code LowCardinality(String),
	visitor_id String
)
ENGINE = Null;

CREATE MATERIALIZED VIEW pageviews_mv TO pageviews AS
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
  FROM events_pageviews

