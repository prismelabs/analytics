CREATE TABLE events_identify(
  timestamp DateTime('UTC'),
  visitor_id String,

  session_uuid UUID,

  -- Properties that are set only once.
  -- JSON keys and values string
  initial_keys Array(String),
  initial_values Array(String),

  -- JSON keys and values string
  keys Array(String),
  values Array(String),
)
ENGINE = Null;

CREATE TABLE users_props_agg(
  visitor_id String,

  initial_session_uuid AggregateFunction(argMin, UUID, DateTime('UTC')),
  latest_session_uuid AggregateFunction(argMax, UUID, DateTime('UTC')),

  -- Properties that are set only once.
  -- JSON keys and values string
  initial_keys AggregateFunction(argMin, Array(String), DateTime('UTC')),
  initial_values AggregateFunction(argMin, Array(String), DateTime('UTC')),

  -- JSON keys and values string
  keys AggregateFunction(argMax, Array(String), DateTime('UTC')),
  values AggregateFunction(argMax, Array(String), DateTime('UTC'))
)
ENGINE = AggregatingMergeTree
ORDER BY visitor_id;

CREATE MATERIALIZED VIEW events_identify_mv TO users_props_agg AS
SELECT
  visitor_id,
  argMinState(session_uuid, timestamp) AS initial_session_uuid,
  argMaxState(session_uuid, timestamp) AS latest_session_uuid,
  argMinState(initial_keys, timestamp) AS initial_keys,
  argMinState(initial_values, timestamp) AS initial_values,
  argMaxState(keys, timestamp) AS keys,
  argMaxState(values, timestamp) AS values
FROM events_identify
GROUP BY visitor_id;

CREATE VIEW users_props AS
SELECT
  visitor_id,
  argMinMerge(initial_session_uuid) AS initial_session_uuid,
  UUIDv7ToDateTime(initial_session_uuid) AS initial_session_timestamp,
  argMaxMerge(latest_session_uuid) AS latest_session_uuid,
  UUIDv7ToDateTime(latest_session_uuid) AS latest_session_timestamp,
  argMinMerge(initial_keys) AS initial_keys,
  argMinMerge(initial_values) AS initial_values,
  argMaxMerge(keys) AS keys,
  argMaxMerge(values) AS values
FROM users_props_agg
GROUP BY visitor_id;

-- index 0 means not found (ClickHouse use 1 as first index).
CREATE FUNCTION IF NOT EXISTS user_prop AS (key) ->
  indexOf(initial_keys, key) > 0
    ? initial_values[indexOf(initial_keys, key)]
    : values[indexOf(keys, key)];

