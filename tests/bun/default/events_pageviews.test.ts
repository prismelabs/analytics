import { expect, test } from 'bun:test'
import { PRISME_URL } from './utils'
import { createClient } from '@clickhouse/client-web'
import { TIMESTAMP_REGEX } from './matchers'

const PAGEVIEWS_ENDPOINT = PRISME_URL + '/api/v1/events/pageviews'

test('GET request instead of POST request', async () => {
  const response = await fetch(PAGEVIEWS_ENDPOINT)
  expect(response.status).toBe(404)
})

test('invalid URL in X-Prisme-Referrer header', async () => {
  const response = await fetch(PAGEVIEWS_ENDPOINT, {
    method: "POST",
    headers: {
      "X-Prisme-Referrer": "not an url",
    }
  })
  expect(response.status).toBe(400)
})

test('invalid URL in Referer header', async () => {
  const response = await fetch(PAGEVIEWS_ENDPOINT, {
    method: "POST",
    headers: {
      "Referer": "not an url",
    }
  })
  expect(response.status).toBe(400)
})

test('non registered domain in X-Prisme-Referrer header is rejected', async () => {
  const response = await fetch(PAGEVIEWS_ENDPOINT, {
    method: "POST",
    headers: {
      "X-Prisme-Referrer": "https://example.com/foo?bar=baz#qux",
    }
  })
  expect(response.status).toBe(400)
})

test('non registered domain in Referer header is rejected', async () => {
  const response = await fetch(PAGEVIEWS_ENDPOINT, {
    method: "POST",
    headers: {
      "Referer": "https://example.com/foo?bar=baz#qux",
    }
  })
  expect(response.status).toBe(400)
})

test('valid URL with registered domain in X-Prisme-Referrer header is accepted', async () => {
  const response = await fetch(PAGEVIEWS_ENDPOINT, {
    method: "POST",
    headers: {
      "X-Prisme-Referrer": "http://mywebsite.localhost/foo?bar=baz#qux"
    }
  })
  expect(response.status).toBe(200)

  // Wait for clickhouse to ingest batch.
  Bun.sleepSync(1000)

  const client = createClient({
    host: "http://clickhouse.localhost:8123",
    username: "clickhouse",
    password: "password",
    database: "prisme",
  })

  const rows = await client.query({
    query: `SELECT * FROM prisme.events_pageviews ORDER BY timestamp DESC LIMIT 1;`
  })
  const data = await rows.json().then((r: any) => r.data[0])

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: "mywebsite.localhost",
    path: "/foo",
    operating_system: "Other",
    browser_family: "Other",
    device: "Other",
  })
})

test('valid URL with registered domain in Referer header is accepted', async () => {
  const response = await fetch(PAGEVIEWS_ENDPOINT, {
    method: "POST",
    headers: {
      "Referer": "http://foo.mywebsite.localhost/another/foo?bar=baz#qux"
    }
  })
  expect(response.status).toBe(200)

  // Wait for clickhouse to ingest batch.
  Bun.sleepSync(1000)

  const client = createClient({
    host: "http://clickhouse.localhost:8123",
    username: "clickhouse",
    password: "password",
    database: "prisme",
  })

  const rows = await client.query({
    query: `SELECT * FROM prisme.events_pageviews ORDER BY timestamp DESC LIMIT 1;`
  })
  const data = await rows.json().then((r: any) => r.data[0])

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: "foo.mywebsite.localhost",
    path: "/another/foo",
    operating_system: "Other",
    browser_family: "Other",
    device: "Other",
  })
})

test('valid pageview with Windows + Chrome user agent', async () => {
  const response = await fetch(PAGEVIEWS_ENDPOINT, {
    method: "POST",
    headers: {
      "Referer": "http://foo.mywebsite.localhost/another/foo?bar=baz#qux",
      "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.3"
    }
  })
  expect(response.status).toBe(200)

  // Wait for clickhouse to ingest batch.
  Bun.sleepSync(1000)

  const client = createClient({
    host: "http://clickhouse.localhost:8123",
    username: "clickhouse",
    password: "password",
    database: "prisme",
  })

  const rows = await client.query({
    query: `SELECT * FROM prisme.events_pageviews ORDER BY timestamp DESC LIMIT 1;`
  })
  const data = await rows.json().then((r: any) => r.data[0])

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: "foo.mywebsite.localhost",
    path: "/another/foo",
    operating_system: "Windows",
    browser_family: "Chrome",
    device: "Other",
  })
})
