ALTER TABLE events_pageviews ADD COLUMN country_code LowCardinality(String) DEFAULT 'XX';
