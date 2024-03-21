RENAME TABLE events_pageviews TO events_pageviews_old;

-- Create a more optimized version of events_pageviews table.
-- This version use a better order by clause and adds a partition by clause
-- that can't be added to an existing table.

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
	country_code LowCardinality(String)
)
ENGINE = MergeTree
PRIMARY KEY (domain, path)
ORDER BY (domain, path, toDate(timestamp), operating_system, browser_family, device, referrer_domain, country_code)
PARTITION BY toYYYYMM(timestamp);

-- Move rows to new table.
INSERT INTO events_pageviews SELECT * FROM events_pageviews_old;

-- Delete old table.
DROP TABLE events_pageviews_old;
