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
    session_id,
  FROM events_pageviews
  WHERE referrer_domain != domain;

-- Drop RocksDB table for direct join as it doesn't work when entry pages and
-- pageviews rows are the same batch.
DROP TABLE entry_pages_datetime;
DROP TABLE entry_pages_datetimes_mv;

-- Exit pageviews.
RENAME TABLE exit_pages_no_bounce TO exit_pageviews_no_bounce;
DROP TABLE exit_pages_mv;
CREATE MATERIALIZED VIEW exit_pageviews_no_bounce_mv TO exit_pageviews_no_bounce AS
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
    session_id,
    entry_timestamp,
  FROM events_pageviews
  WHERE referrer_domain == domain;

