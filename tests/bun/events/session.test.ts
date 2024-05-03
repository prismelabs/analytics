import { expect, test } from 'bun:test'
import { faker } from '@faker-js/faker'

import { createClient } from '@clickhouse/client-web'
import { COUNTRY_CODE_REGEX, PRISME_PAGEVIEWS_URL, PRISME_VISITOR_ID_REGEX, SESSION_ID_REGEX, TIMESTAMP_REGEX } from '../const'

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

  // Fetch latest exit and entry page.
  let exitPage = await getLatestExitPage()
  let entryPage = await getLatestEntryPage()

  // Session ID must not match.
  if (exitPage !== undefined) { expect(exitPage.session_id).not.toBe(entryPage.session_id) }

  // Check entry page.
  const entryPageMatcher = {
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: 'foo.mywebsite.localhost',
    path: '/page1',
    operating_system: 'Other',
    browser_family: 'Other',
    device: 'Other',
    referrer_domain: 'direct',
    country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
    visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
    session_id: expect.stringMatching(SESSION_ID_REGEX)
  }
  expect(entryPage).toMatchObject(entryPageMatcher)

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

  exitPage = await getLatestExitPage()
  entryPage = await getLatestEntryPage()

  // Exit page was added.
  expect(exitPage).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: entryPage.domain,
    path: '/page2',
    operating_system: entryPage.operating_system,
    browser_family: entryPage.browser_family,
    device: entryPage.device,
    referrer_domain: entryPage.domain,
    country_code: entryPage.country_code,
    visitor_id: entryPage.visitor_id,
    session_id: entryPage.session_id,
    entry_timestamp: entryPage.timestamp
  })

  // Entry page remains the same.
  expect(entryPage).toMatchObject(entryPageMatcher)

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

  exitPage = await getLatestExitPage()
  entryPage = await getLatestEntryPage()

  // New exit page replace /page2.
  const exitPageMatcher = {
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: entryPage.domain,
    path: '/page3',
    operating_system: entryPage.operating_system,
    browser_family: entryPage.browser_family,
    device: entryPage.device,
    referrer_domain: entryPage.domain,
    country_code: entryPage.country_code,
    visitor_id: entryPage.visitor_id,
    session_id: entryPage.session_id,
    entry_timestamp: entryPage.timestamp
  }
  expect(exitPage).toMatchObject(exitPageMatcher)

  // Entry page remains the same.
  expect(entryPage).toMatchObject(entryPageMatcher)

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

  exitPage = await getLatestExitPage()
  entryPage = await getLatestEntryPage()

  // Exit page remains unchanged.
  expect(exitPage).toMatchObject(exitPageMatcher)

  // New entry page.
  entryPageMatcher.path = '/page4'
  expect(entryPage).toMatchObject(entryPageMatcher)
}, { timeout: 30_000 })

async function getLatestEntryPage (): Promise<any> {
  // Wait for clickhouse to ingest batch.
  Bun.sleepSync(1000)

  const client = createClient({
    host: 'http://clickhouse.localhost:8123',
    username: 'clickhouse',
    password: 'password',
    database: 'prisme'
  })

  const rows = await client.query({
    query: 'SELECT * FROM prisme.entry_pages ORDER BY timestamp DESC LIMIT 1;'
  })
  return rows.json().then((r: any) => r.data[0])
}

async function getLatestExitPage (): Promise<any> {
  // Wait for clickhouse to ingest batch.
  Bun.sleepSync(1000)

  const client = createClient({
    host: 'http://clickhouse.localhost:8123',
    username: 'clickhouse',
    password: 'password',
    database: 'prisme'
  })

  const rows = await client.query({
    query: 'SELECT * FROM prisme.exit_pages_no_bounce FINAL ORDER BY timestamp DESC LIMIT 1;'
  })
  return rows.json().then((r: any) => r.data[0])
}
