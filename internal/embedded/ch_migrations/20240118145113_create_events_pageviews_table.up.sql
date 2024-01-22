CREATE TABLE IF NOT EXISTS events_pageviews (
	timestamp DateTime('UTC'),
	domain String,
	path String,
	operating_system LowCardinality(String),
	browser_family LowCardinality(String),
	device LowCardinality(String)
)
ENGINE = MergeTree
ORDER BY (timestamp, domain, path);
