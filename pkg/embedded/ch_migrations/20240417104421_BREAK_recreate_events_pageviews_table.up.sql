RENAME TABLE events_pageviews TO events_pageviews_old;

-- Add visitor_id to primary key so it can be used for sampling.

-- It is beneficial for queries to order the primary key columns by cardinality in ascending order.
-- https://clickhouse.com/docs/en/optimize/sparse-primary-indexes#summary
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
ENGINE = MergeTree
PRIMARY KEY (domain, path, toDate(timestamp), xxh3(visitor_id))
ORDER BY (domain, path, toDate(timestamp), xxh3(visitor_id), operating_system, browser_family, device, referrer_domain, country_code)
SAMPLE BY xxh3(visitor_id)
PARTITION BY toYYYYMM(timestamp);

-- Move rows to new table.
INSERT INTO events_pageviews SELECT * FROM events_pageviews_old;

-- Delete old table.
DROP TABLE events_pageviews_old;
