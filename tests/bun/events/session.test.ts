import { expect, test } from 'bun:test'
import { faker } from '@faker-js/faker'

import { createClient } from '@clickhouse/client-web'
import { COUNTRY_CODE_REGEX, PRISME_PAGEVIEWS_URL, PRISME_VISITOR_ID_REGEX, TIMESTAMP_REGEX, UUID_V7_REGEX } from '../const'

test('multiple page view session', async () => {
  const ipAddr = faker.internet.ip()

  let response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': ipAddr,
      'X-Prisme-Referrer': 'http://foo.mywebsite.localhost/page1'
    }
  })
  expect(response.status).toBe(200)

  let session = await getLatestSession()
  const sessionUuid = session.session_uuid
  const visitorId = session.visitor_id
  const countryCode = session.country_code

  expect(session).toMatchObject({
    domain: 'foo.mywebsite.localhost',
    entry_path: '/page1',
    exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    exit_path: '/page1',
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
  })

  // View another page.
  response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': ipAddr,
      'X-Prisme-Document-Referrer': 'http://foo.mywebsite.localhost/page1',
      'X-Prisme-Referrer': 'http://foo.mywebsite.localhost/page2'
    }
  })

  session = await getLatestSession()

  expect(session).toMatchObject({
    domain: 'foo.mywebsite.localhost',
    entry_path: '/page1',
    exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    exit_path: '/page2',
    operating_system: 'Other',
    browser_family: 'Other',
    device: 'Other',
    referrer_domain: 'direct',
    country_code: countryCode,
    visitor_id: visitorId,
    session_uuid: sessionUuid,
    utm_source: '',
    utm_medium: '',
    utm_campaign: '',
    utm_term: '',
    utm_content: '',
    version: 2
  })

  // View a third page.
  response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': ipAddr,
      'X-Prisme-Document-Referrer': 'http://foo.mywebsite.localhost/page2',
      'X-Prisme-Referrer': 'http://foo.mywebsite.localhost/page3'
    }
  })

  session = await getLatestSession()

  expect(session).toMatchObject({
    domain: 'foo.mywebsite.localhost',
    entry_path: '/page1',
    exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    exit_path: '/page3',
    operating_system: 'Other',
    browser_family: 'Other',
    device: 'Other',
    referrer_domain: 'direct',
    country_code: countryCode,
    visitor_id: visitorId,
    session_uuid: sessionUuid,
    utm_source: '',
    utm_medium: '',
    utm_campaign: '',
    utm_term: '',
    utm_content: '',
    version: 3
  })

  // Start a new session!
  response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': ipAddr,
      'X-Prisme-Referrer': 'http://foo.mywebsite.localhost/page4'
      // No document referrer.
    }
  })
  expect(response.status).toBe(200)

  session = await getLatestSession()

  expect(session.sessionUuid).not.toBe(sessionUuid)
  expect(session).toMatchObject({
    domain: 'foo.mywebsite.localhost',
    entry_path: '/page4',
    exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    exit_path: '/page4',
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
  })
}, { timeout: 30_000 })

async function getLatestSession (): Promise<any> {
  // Wait for clickhouse to ingest batch.
  Bun.sleepSync(1000)

  const client = createClient({
    host: 'http://clickhouse.localhost:8123',
    username: 'clickhouse',
    password: 'password',
    database: 'prisme'
  })

  const rows = await client.query({
    query: 'SELECT * FROM prisme.sessions WHERE sign = 1 ORDER BY exit_timestamp DESC LIMIT 1;'
  })
  const session = await rows.json().then((r: any) => r.data[0])
  delete session.sign

  return session
}
