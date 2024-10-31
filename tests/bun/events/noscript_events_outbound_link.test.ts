import { expect, test } from 'bun:test'
import { faker } from '@faker-js/faker'

import { createClient } from '@clickhouse/client-web'
import { COUNTRY_CODE_REGEX, PRISME_NOSCRIPT_OUTBOUND_LINK_EVENTS_URL, PRISME_PAGEVIEWS_URL, PRISME_VISITOR_ID_REGEX, TIMESTAMP_REGEX, UUID_V7_REGEX } from '../const'
import { randomIpWithSession } from '../utils'

const seed = new Date().getTime()
console.log('faker seed', seed)
faker.seed(seed)

test('POST request instead of GET request', async () => {
  const response = await fetch(PRISME_NOSCRIPT_OUTBOUND_LINK_EVENTS_URL + '?url=http://www.example.com', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      Referer: 'http://mywebsite.localhost/foo'
    }
  })
  expect(response.status).toBe(405)
})

test('invalid URL in Referer header', async () => {
  const response = await fetch(PRISME_NOSCRIPT_OUTBOUND_LINK_EVENTS_URL + '?url=http://www.example.com', {
    method: 'GET',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      Referer: 'not an url'
    }
  })
  expect(response.status).toBe(400)
})

test('non registered domain in Origin header is rejected', async () => {
  const response = await fetch(PRISME_NOSCRIPT_OUTBOUND_LINK_EVENTS_URL + '?url=http://www.example.com', {
    method: 'GET',
    headers: {
      Origin: 'https://example.com',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      Referer: 'https://example.com/foo?bar=baz#qux'
    }
  })
  expect(response.status).toBe(400)
})

test('invalid sessionless custom event', async () => {
  const response = await fetch(PRISME_NOSCRIPT_OUTBOUND_LINK_EVENTS_URL + '?url=http://www.example.com', {
    method: 'GET',
    headers: {
      Origin: 'http://mywebsite.localhost',
      // No session associated with this ip.
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'http://mywebsite.localhost/index.html'
    }
  })

  expect(response.status).toBe(400)
})

test('invalid url query param with relative url', async () => {
  const response = await fetch(PRISME_NOSCRIPT_OUTBOUND_LINK_EVENTS_URL + '?url=/foo/bar', {
    method: 'GET',
    headers: {
      Origin: 'https://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      Referer: 'https://mywebsite.localhost/'
    }
  })

  expect(response.status).toBe(400)
})

test('valid test cases break', async () => {
  // Sleep so pageviews and custom timestamps are different for valid test
  // cases.
  // Without this sleep, getLatestXXX function may return rows from invalid test
  // cases.
  // This is not needed later as each getLatestXXX contains a 1s sleep.
  Bun.sleepSync(1000)
})

test('valid outbound link click event', async () => {
  const response = await fetch(PRISME_NOSCRIPT_OUTBOUND_LINK_EVENTS_URL + '?url=https://www.example.com/page1', {
    method: 'GET',
    headers: {
      Origin: 'https://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      Referer: 'https://mywebsite.localhost'
    },
    redirect: 'manual'
  })
  expect(response.status).toBe(302)
  expect(response.headers.get('Location')).toBe('https://www.example.com/page1')

  const data = await getLatestOutboundLinkClickEvent()

  expect(data).toMatchObject({
    session: {
      domain: 'mywebsite.localhost',
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
    event: {
      domain: 'mywebsite.localhost',
      path: '/',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      link: 'https://www.example.com/page1'
    }
  })
})

test('concurrent pageview and outbound link click events', async () => {
  const ipAddr = faker.internet.ip()

  await Promise.all([
    // Custom events first.
    fetch(PRISME_NOSCRIPT_OUTBOUND_LINK_EVENTS_URL + '?url=https://www.example.com/page1', {
      method: 'GET',
      headers: {
        Origin: 'https://mywebsite.localhost',
        'X-Forwarded-For': ipAddr,
        Referer: 'https://mywebsite.localhost'
      },
      redirect: 'manual'
    }),

    // Pageview concurrently.
    // This pageview will create session that will be used for both events.
    fetch(PRISME_PAGEVIEWS_URL, {
      method: 'POST',
      headers: {
        Origin: 'https://mywebsite.localhost',
        'X-Forwarded-For': ipAddr,
        'X-Prisme-Referrer': 'https://mywebsite.localhost/'
      }
    })
  ]).then((results) => results.forEach((resp) => expect(resp.status).toBeLessThan(400)))

  const data = await getLatestOutboundLinkClickEvent()

  expect(data).toMatchObject({
    session: {
      domain: 'mywebsite.localhost',
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
    event: {
      domain: 'mywebsite.localhost',
      path: '/',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      link: 'https://www.example.com/page1'
    }
  })
})

test('valid outbound link click event with complete referrer', async () => {
  const response = await fetch(PRISME_NOSCRIPT_OUTBOUND_LINK_EVENTS_URL + '?url=https://www.example.com/page1', {
    method: 'GET',
    headers: {
      Origin: 'https://mywebsite.localhost',
      // Create session on / to emulate origin referrer policy of pageview event.
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost', { path: '/' }),
      // Referrer contains path to emulate <a referrerpolicy="unsafe-url">
      Referer: 'https://mywebsite.localhost/foo'
    },
    redirect: 'manual'
  })
  expect(response.status).toBe(302)
  expect(response.headers.get('Location')).toBe('https://www.example.com/page1')

  const data = await getLatestOutboundLinkClickEvent()

  expect(data).toMatchObject({
    session: {
      domain: 'mywebsite.localhost',
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
    event: {
      domain: 'mywebsite.localhost',
      path: '/',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      link: 'https://www.example.com/page1'
    }
  })
})

test('valid outbound link click event with referrer in query', async () => {
  const response = await fetch(PRISME_NOSCRIPT_OUTBOUND_LINK_EVENTS_URL + '?url=https://www.example.com/page1&referrer=https://mywebsite.localhost/foo', {
    method: 'GET',
    headers: {
      Origin: 'https://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost')
    },
    redirect: 'manual'
  })
  expect(response.status).toBe(302)
  expect(response.headers.get('Location')).toBe('https://www.example.com/page1')

  const data = await getLatestOutboundLinkClickEvent()

  expect(data).toMatchObject({
    session: {
      domain: 'mywebsite.localhost',
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
    event: {
      domain: 'mywebsite.localhost',
      path: '/', // not /foo as session current page is /
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      link: 'https://www.example.com/page1'
    }
  })
})

async function getLatestOutboundLinkClickEvent (): Promise<any> {
  // Wait for clickhouse to ingest batch.
  Bun.sleepSync(1000)

  const client = createClient({
    host: 'http://clickhouse.localhost:8123',
    username: 'clickhouse',
    password: 'password',
    database: 'prisme'
  })

  const sessions = await client.query({
    query: 'SELECT * FROM sessions ORDER BY exit_timestamp DESC LIMIT 1'
  })
  const session = await sessions.json().then((r: any) => r.data[0])
  expect(session.sign).toBe(1)
  delete session.sign

  const clickEvents = await client.query({
    query: `SELECT * FROM outbound_link_clicks WHERE visitor_id = '${session.visitor_id as string}' ORDER BY timestamp DESC LIMIT 1`
  })
  const clickEvent = await clickEvents.json().then((r: any) => r.data[0])
  if (clickEvent === null || clickEvent === undefined) return null
  expect(clickEvent.visitor_id).toBe(session.visitor_id)
  expect(clickEvent.session_uuid).toBe(session.session_uuid)

  return { event: clickEvent, session }
}
