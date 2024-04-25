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
	operating_system LowCardinality(String),
	browser_family LowCardinality(String),
	device LowCardinality(String),
	referrer_domain String,
	country_code LowCardinality(String),
	visitor_id String,
	-- event name
	name String,
	-- JSON keys and values string
	keys Array(String),
	values Array(String)
)
ENGINE = MergeTree
PRIMARY KEY (domain, name, path, toDate(timestamp), xxh3(visitor_id))
ORDER BY (domain, name, path, toDate(timestamp), xxh3(visitor_id), operating_system, browser_family, device, referrer_domain, country_code)
SAMPLE BY xxh3(visitor_id)
PARTITION BY toYYYYMM(timestamp);

-- Move rows to new table.
INSERT INTO events_custom 
	SELECT timestamp, domain, path, operating_system, browser_family, device,
		referrer_domain, country_code, visitor_id, name, keys, values
	FROM events_custom_old;

-- Delete old table.
DROP TABLE events_custom_old;
