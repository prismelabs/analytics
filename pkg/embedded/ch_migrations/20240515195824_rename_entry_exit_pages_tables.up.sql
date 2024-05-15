-- Entry pageviews.
RENAME TABLE entry_pages TO entry_pageviews;
DROP TABLE entry_pages_mv;
CREATE MATERIALIZED VIEW entry_pageviews_mv TO entry_pageviews AS
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
    xxh3(timestamp, visitor_id, domain) AS session_id
  FROM events_pageviews
  WHERE referrer_domain != domain;

-- RocksDB for direct join.
RENAME TABLE entry_pages_datetime to entry_pageviews_rocksdb;
DROP TABLE entry_pages_datetimes_mv;
CREATE MATERIALIZED VIEW entry_pageviews_rocksdb_mv TO entry_pageviews_rocksdb AS
  SELECT
    timestamp,
    domain,
    visitor_id,
    session_id
  FROM entry_pageviews;

-- Exit pageviews.
RENAME TABLE exit_pages_no_bounce TO exit_pageviews_no_bounce;
DROP TABLE exit_pages_mv;
CREATE MATERIALIZED VIEW exit_pageviews_no_bounce_mv TO exit_pageviews_no_bounce AS
  SELECT
    events_pageviews.timestamp,
    events_pageviews.domain,
    events_pageviews.path,
    events_pageviews.operating_system,
    events_pageviews.browser_family,
    events_pageviews.device,
    events_pageviews.referrer_domain,
    events_pageviews.country_code,
    events_pageviews.visitor_id,
    entry_pageviews_rocksdb.timestamp AS entry_timestamp,
    entry_pageviews_rocksdb.session_id
  FROM events_pageviews
  INNER JOIN entry_pageviews_rocksdb
  ON entry_pageviews_rocksdb.visitor_id = events_pageviews.visitor_id
  WHERE referrer_domain == domain;

