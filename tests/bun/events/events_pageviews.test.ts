import { expect, test } from 'bun:test'
import { faker } from '@faker-js/faker'

import { createClient } from '@clickhouse/client-web'
import { COUNTRY_CODE_REGEX, PRISME_PAGEVIEWS_URL, PRISME_VISITOR_ID_REGEX, TIMESTAMP_REGEX } from '../const'

const seed = new Date().getTime()
console.log('faker seed', seed)
faker.seed(seed)

test('GET request instead of POST request', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/foo'
    }
  })
  expect(response.status).toBe(405)
})

test('invalid URL in X-Prisme-Referrer header', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'not an url'
    }
  })
  expect(response.status).toBe(400)
})

test('invalid URL in X-Prisme-Document-Referrer header', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'http://mywebsite.localhost',
      'X-Prisme-Document-Referrer': 'not an url'
    }
  })
  expect(response.status).toBe(400)
})

test('invalid URL in Referer header', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'not an url'
    }
  })
  expect(response.status).toBe(400)
})

test('non registered domain in Origin header is rejected', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'https://example.com',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'https://example.com/foo?bar=baz#qux'
    }
  })
  expect(response.status).toBe(400)
})

test('registered domain in Origin header and valid referrer is acepted is accepted', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'https://example.com/foo?bar=baz#qux'
    }
  })
  expect(response.status).toBe(200)
})

test('registered domain in Origin header and valid X-Prisme-Referrer is accepted', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
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
    referrer_domain: 'www.example.com',
    country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
    visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX)
  })
})

test('valid URL with registered domain in Origin header is accepted', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
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
    device: 'Other',
    referrer_domain: 'www.example.com',
    country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
    visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX)
  })
})

test('valid pageview with Windows + Chrome user agent', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
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
    referrer_domain: 'www.example.com',
    country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
    visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX)
  })
})

test('valid pageview without X-Prisme-Document-Referrer', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
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
    referrer_domain: 'direct',
    country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
    visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX)
  })
})

test('valid pageview without trailing slash in referrer', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'http://foo.mywebsite.localhost' // No / after localhost
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: 'foo.mywebsite.localhost',
    path: '/', // path contains /
    operating_system: 'Other',
    browser_family: 'Other',
    device: 'Other',
    referrer_domain: 'direct',
    country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
    visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX)
  })
})

test('valid pageview with US IP address', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': '8.8.8.8', // Google public DNS
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
    referrer_domain: 'direct',
    country_code: 'US',
    visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX)
  })
})

test('valid pageview with dirty path', async () => {
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': '8.8.8.8', // Google public DNS
      Referer: 'http://foo.mywebsite.localhost///another/../another/foo?bar=baz#qux'
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
    referrer_domain: 'direct',
    country_code: 'US',
    visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX)
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
  return await rows.json().then((r: any) => r.data[0])
}
