CREATE TABLE events_custom (
  timestamp DateTime('UTC'),
  domain String,
  path String,
  name String,
  -- Simple JSON string
  properties String,
)
ENGINE = MergeTree
ORDER BY (domain, name);

