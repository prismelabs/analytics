-- It is beneficial for queries to order the primary key columns by cardinality in ascending order.
-- https://clickhouse.com/docs/en/optimize/sparse-primary-indexes#summary
CREATE TABLE events_custom (
	-- Same attributes as page views.
	timestamp DateTime('UTC'),
	domain String,
	path String,
	operating_system LowCardinality(String),
	browser_family LowCardinality(String),
	device LowCardinality(String),
	referrer_domain String,
	country_code LowCardinality(String),
	-- event name
	name String,
	-- JSON keys and values string
	keys Array(String),
	values Array(String)
)
ENGINE = MergeTree
PRIMARY KEY (domain, name)
ORDER BY (domain, name, path, toDate(timestamp), operating_system, browser_family, device, referrer_domain, country_code)
PARTITION BY toYYYYMM(timestamp);

CREATE FUNCTION IF NOT EXISTS event_property AS (key) -> values[indexOf(keys, key)];

