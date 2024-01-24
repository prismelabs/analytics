import { expect, test } from 'bun:test'
import { faker } from '@faker-js/faker'

import { createClient } from '@clickhouse/client-web'
import { PRISME_PAGEVIEWS_URL, TIMESTAMP_REGEX } from '../const'

const seed = new Date().getTime()
console.log('faker seed', seed)
faker.seed(seed)

test('GET request instead of POST request', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL)
  expect(response.status).toBe(404)
})

test('invalid URL in X-Prisme-Referrer header', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'not an url'
    }
  })
  expect(response.status).toBe(400)
})

test('invalid URL in Referer header', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'not an url'
    }
  })
  expect(response.status).toBe(400)
})

test('non registered domain in X-Prisme-Referrer header is rejected', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'https://example.com/foo?bar=baz#qux'
    }
  })
  expect(response.status).toBe(400)
})

test('non registered domain in Referer header is rejected', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'https://example.com/foo?bar=baz#qux'
    }
  })
  expect(response.status).toBe(400)
})

test('valid URL with registered domain in X-Prisme-Referrer header is accepted', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/foo?bar=baz#qux',
      'X-Prisme-Document-Referrer': 'https://www.example.com/foo'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: 'mywebsite.localhost',
    path: '/foo',
    operating_system: 'Other',
    browser_family: 'Other',
    device: 'Other',
    referrer_domain: 'www.example.com'
  })
})

test('valid URL with registered domain in Referer header is accepted', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'http://foo.mywebsite.localhost/another/foo?bar=baz#qux',
      'X-Prisme-Document-Referrer': 'https://www.example.com/foo'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: 'foo.mywebsite.localhost',
    path: '/another/foo',
    operating_system: 'Other',
    browser_family: 'Other',
    device: 'Other'
  })
})

test('valid pageview with Windows + Chrome user agent', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'http://foo.mywebsite.localhost/another/foo?bar=baz#qux',
      'X-Prisme-Document-Referrer': 'https://www.example.com/foo',
      'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.3'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: 'foo.mywebsite.localhost',
    path: '/another/foo',
    operating_system: 'Windows',
    browser_family: 'Chrome',
    device: 'Other',
    referrer_domain: 'www.example.com'
  })
})

test('valid pageview without X-Prisme-Document-Referrer', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'http://foo.mywebsite.localhost/another/foo?bar=baz#qux'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: 'foo.mywebsite.localhost',
    path: '/another/foo',
    operating_system: 'Other',
    browser_family: 'Other',
    device: 'Other',
    referrer_domain: 'direct'
  })
})

async function getLatestPageview (): Promise<any> {
  // Wait for clickhouse to ingest batch.
  Bun.sleepSync(1000)

  const client = createClient({
    host: 'http://clickhouse.localhost:8123',
    username: 'clickhouse',
    password: 'password',
    database: 'prisme'
  })

  const rows = await client.query({
    query: 'SELECT * FROM prisme.events_pageviews ORDER BY timestamp DESC LIMIT 1;'
  })
  return rows.json().then((r: any) => r.data[0])
}
