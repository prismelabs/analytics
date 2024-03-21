-- It is beneficial for queries to order the primary key columns by cardinality in ascending order.
-- https://clickhouse.com/docs/en/optimize/sparse-primary-indexes#summary
CREATE TABLE events_custom (
  timestamp DateTime('UTC'),
  domain String,
  path String,
  name String,
  -- JSON keys and values string
	keys Array(String),
	values Array(String)
)
ENGINE = MergeTree
PRIMARY KEY (domain, name)
ORDER BY (domain, name, path, toDate(timestamp))
PARTITION BY toYYYYMM(timestamp);

CREATE FUNCTION event_property AS (key) -> values[indexOf(keys, key)];

