export const PRISME_URL = 'http://prisme.localhost'
export const PRISME_ADMIN_URL = 'http://prisme.localhost:9090'

export const PRISME_API_URL = PRISME_URL + '/api/v1'
export const PRISME_PAGEVIEWS_URL = PRISME_API_URL + '/events/pageviews'
export const PRISME_NOSCRIPT_PAGEVIEWS_URL = PRISME_API_URL + '/noscript/events/pageviews'
export const PRISME_CUSTOM_EVENTS_URL = PRISME_API_URL + '/events/custom'
export const PRISME_NOSCRIPT_CUSTOM_EVENTS_URL = PRISME_API_URL + '/noscript/events/custom'
export const PRISME_IDENTIFY_EVENTS_URL = PRISME_API_URL + '/events/identify'
export const PRISME_NOSCRIPT_IDENTIFY_EVENTS_URL = PRISME_API_URL + '/noscript/events/identify'

export const PRISME_METRICS_URL = PRISME_ADMIN_URL + '/metrics'

export const TIMESTAMP_REGEX = /\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}/
export const UUID_V4_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i
export const UUID_V7_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i
export const COUNTRY_CODE_REGEX = /^[A-Z]{2}$/
export const PRISME_VISITOR_ID_REGEX = /prisme_.+/
