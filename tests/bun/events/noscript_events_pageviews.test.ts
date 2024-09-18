import { expect, test } from 'bun:test'
import { faker } from '@faker-js/faker'

import { createClient } from '@clickhouse/client-web'
import { COUNTRY_CODE_REGEX, PRISME_NOSCRIPT_PAGEVIEWS_URL, PRISME_VISITOR_ID_REGEX, TIMESTAMP_REGEX, UUID_V7_REGEX } from '../const'

const seed = new Date().getTime()
console.log('faker seed', seed)
faker.seed(seed)

test('POST request instead of GET request', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'http://mywebsite.localhost/foo'
    }
  })
  expect(response.status).toBe(405)
})

test('invalid URL in Referer header', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'not an url'
    }
  })
  expect(response.status).toBe(400)
})

test('invalid URL in X-Prisme-Document-Referrer header', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL + '?document-referrer=not an url', {
    method: 'GET',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'http://mywebsite.localhost'
    }
  })
  expect(response.status).toBe(400)
})

test('non registered domain in Origin header is rejected', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'https://example.com',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'https://example.com/foo?bar=baz#qux'
    }
  })
  expect(response.status).toBe(400)
})

test('robot user agent is rejected', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'https://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'https://mywebsite.localhost/foo?bar=baz#qux',
      'User-Agent': 'Googlebot'
    }
  })
  expect(response.status).toBe(400)
})

test('valid test cases break', async () => {
  // Sleep so pageviews and identify timestamps are different for valid test
  // cases.
  // Without this sleep, getLatestXXX function may return rows from invalid test
  // cases.
  // This is not needed later as each getLatestXXX contains a 1s sleep.
  Bun.sleepSync(1000)
})

test('valid internal pageview with no session associated', async () => {
  // internal traffic, but no session exist, pageview should be rejected.
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL + '?document-referrer=https://mywebsite.localhost/foo?bar=baz#qux', {
    method: 'GET',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'https://mywebsite.localhost/bar?bar=baz#qux'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'mywebsite.localhost',
      entry_path: '/bar',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/bar',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'mywebsite.localhost',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'mywebsite.localhost',
      path: '/bar',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('valid pageview with different referrer query param and Referer header', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL + '?referrer=https://bar.mywebsite.localhost/bar?bar=baz#qux', {
    method: 'GET',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'https://foo.mywebsite.localhost/foo?bar=baz#qux'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'bar.mywebsite.localhost',
      entry_path: '/bar',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/bar',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'direct',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'bar.mywebsite.localhost',
      path: '/bar',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('valid pageview with different document-referrer query param and X-Prisme-Document-Referrer header', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL + '?document-referrer=https://bar.mywebsite.localhost/bar?bar=baz#qux', {
    method: 'GET',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'https://mywebsite.localhost/',
      'X-Prisme-Document-Referrer': 'https://foo.mywebsite.localhost/foo?bar=baz#qux'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'mywebsite.localhost',
      entry_path: '/',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'bar.mywebsite.localhost',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'mywebsite.localhost',
      path: '/',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('registered domain in Origin header and valid referrer is accepted', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'https://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'https://mywebsite.localhost/foo?bar=baz#qux'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'mywebsite.localhost',
      entry_path: '/foo',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/foo',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'direct',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'mywebsite.localhost',
      path: '/foo',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('registered domain in Origin header and valid Referer is accepted', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL + '?document-referrer=https://www.example.com/foo', {
    method: 'GET',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'http://mywebsite.localhost/foo?bar=baz#qux'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'mywebsite.localhost',
      entry_path: '/foo',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/foo',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'www.example.com',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'mywebsite.localhost',
      path: '/foo',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('valid URL with registered domain in Origin header is accepted', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL + '?document-referrer=https://www.example.com/foo', {
    method: 'GET',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'http://foo.mywebsite.localhost/another/foo?bar=baz#qux'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'foo.mywebsite.localhost',
      entry_path: '/another/foo',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/another/foo',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'www.example.com',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'foo.mywebsite.localhost',
      path: '/another/foo',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('valid pageview with Windows + Chrome user agent', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL + '?document-referrer=https://www.example.com/foo', {
    method: 'GET',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'http://foo.mywebsite.localhost/another/foo?bar=baz#qux',
      'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.3'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'foo.mywebsite.localhost',
      entry_path: '/another/foo',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/another/foo',
      operating_system: 'Windows',
      browser_family: 'Chrome',
      device: 'Other',
      referrer_domain: 'www.example.com',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'foo.mywebsite.localhost',
      path: '/another/foo',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('valid pageview without X-Prisme-Document-Referrer', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'http://foo.mywebsite.localhost/another/foo?bar=baz#qux'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'foo.mywebsite.localhost',
      entry_path: '/another/foo',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/another/foo',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'direct',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'foo.mywebsite.localhost',
      path: '/another/foo',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('valid pageview without trailing slash in referrer', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'http://foo.mywebsite.localhost' // No / after localhost
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'foo.mywebsite.localhost',
      entry_path: '/', // path contains /
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/', // path contains /
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'direct',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'foo.mywebsite.localhost',
      path: '/',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('valid pageview with US IP address', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': '8.8.8.8', // Google public DNS
      Referer: 'http://foo.mywebsite.localhost/us/foo2?bar=baz#qux'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'foo.mywebsite.localhost',
      entry_path: '/us/foo2',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/us/foo2',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'direct',
      country_code: 'US',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'foo.mywebsite.localhost',
      path: '/us/foo2',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('valid pageview with dirty path', async () => {
  const ipAddr = faker.internet.ip()
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': ipAddr,
      Referer: 'http://foo.mywebsite.localhost///another/../another/foo?bar=baz#qux'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'foo.mywebsite.localhost',
      entry_path: '/another/foo',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/another/foo',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'direct',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'foo.mywebsite.localhost',
      path: '/another/foo',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('valid pageview with UTM parameters', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      // utm_id extra query param should be ignored.
      Referer: 'http://foo.mywebsite.localhost?utm_source=github&utm_medium=ppc&utm_campaign=spring+sale&utm_id=aa&utm_term=running+shoes&utm_content=logolink'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'foo.mywebsite.localhost',
      entry_path: '/', // path contains /
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/', // path contains /
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'direct',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: 'github',
      utm_medium: 'ppc',
      utm_campaign: 'spring sale',
      utm_term: 'running shoes',
      utm_content: 'logolink',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'foo.mywebsite.localhost',
      path: '/',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('valid pageview with ref query parameter', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'http://foo.mywebsite.localhost/?ref=advertising1'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'foo.mywebsite.localhost',
      entry_path: '/',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'direct',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: 'advertising1',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'foo.mywebsite.localhost',
      path: '/',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('valid consecutive pageviews', async () => {
  const ipAddr = faker.internet.ip()
  let response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': ipAddr,
      Referer: 'http://foo.mywebsite.localhost/'
    }
  })
  expect(response.status).toBe(200)

  let data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'foo.mywebsite.localhost',
      entry_path: '/',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'direct',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'foo.mywebsite.localhost',
      path: '/',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })

  response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL + '?document-referrer=http://foo.mywebsite.localhost/', {
    method: 'GET',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': ipAddr,
      Referer: 'http://foo.mywebsite.localhost/foo'
    }
  })
  expect(response.status).toBe(200)

  data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'foo.mywebsite.localhost',
      entry_path: '/',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/foo',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'direct',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 2
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'foo.mywebsite.localhost',
      path: '/foo',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('valid pageview with custom visitor id', async () => {
  const visitorId = `visitor-id-${Math.random()}`
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL + `?visitor-id=${visitorId}`, {
    method: 'GET',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Visitor-Id': visitorId,
      Referer: 'http://foo.mywebsite.localhost/'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'foo.mywebsite.localhost',
      entry_path: '/',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'direct',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: visitorId,
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'foo.mywebsite.localhost',
      path: '/',
      visitor_id: visitorId,
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('valid pageview with empty visitor id fallback to auto generated visitor id', async () => {
  const response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Visitor-Id': '',
      Referer: 'http://foo.mywebsite.localhost/'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'foo.mywebsite.localhost',
      entry_path: '/',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'direct',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'foo.mywebsite.localhost',
      path: '/',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })
})

test('valid consecutive pageviews with visitor id defined on second event', async () => {
  const ipAddr = faker.internet.ip()
  let response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': ipAddr,
      Referer: 'http://foo.mywebsite.localhost/'
    }
  })
  expect(response.status).toBe(200)

  let data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'foo.mywebsite.localhost',
      entry_path: '/',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'direct',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 1
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'foo.mywebsite.localhost',
      path: '/',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
  })

  const visitorId = `visitor-id-${Math.random()}`
  response = await fetch(PRISME_NOSCRIPT_PAGEVIEWS_URL + `?visitor-id=${visitorId}&document-referrer=http://foo.mywebsite.localhost/`, {
    method: 'GET',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': ipAddr,
      Referer: 'http://foo.mywebsite.localhost/foo'
    }
  })
  expect(response.status).toBe(200)

  data = await getLatestPageview()

  expect(data).toMatchObject({
    session: {
      domain: 'foo.mywebsite.localhost',
      entry_path: '/',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/foo',
      operating_system: 'Other',
      browser_family: 'Other',
      device: 'Other',
      referrer_domain: 'direct',
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: visitorId,
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_term: '',
      utm_content: '',
      version: 2
    },
    pageview: {
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'foo.mywebsite.localhost',
      path: '/foo',
      visitor_id: visitorId,
      session_uuid: expect.stringMatching(UUID_V7_REGEX)
    }
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

  const sessions = await client.query({
    query: 'SELECT * FROM sessions WHERE sign = 1 ORDER BY exit_timestamp DESC LIMIT 1'
  })
  const session = await sessions.json().then((r: any) => r.data[0])
  delete session.sign

  const pageviews = await client.query({
    query: 'SELECT * FROM prisme.pageviews ORDER BY timestamp DESC LIMIT 1;'
  })
  const pageview = await pageviews.json().then((r: any) => r.data[0])
  expect(pageview.session_uuid).toBe(session.session_uuid)
  expect(pageview.visitor_id).toBe(session.visitor_id)

  return { session, pageview }
}
