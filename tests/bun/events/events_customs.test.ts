import { expect, test } from 'bun:test'
import { faker } from '@faker-js/faker'

import { createClient } from '@clickhouse/client-web'
import { COUNTRY_CODE_REGEX, PRISME_CUSTOM_EVENTS_URL, PRISME_PAGEVIEWS_URL, PRISME_VISITOR_ID_REGEX, TIMESTAMP_REGEX, UUID_V7_REGEX } from '../const'
import { randomIpWithSession } from '../utils'

const seed = new Date().getTime()
console.log('faker seed', seed)
faker.seed(seed)

test('GET request instead of POST request', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'GET',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/',
      'Content-Type': 'application/json'
    }
    // body: JSON.stringify({}) // GET request can't have body.
  })
  expect(response.status).toBe(405)
})

test('invalid URL in X-Prisme-Referrer header', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      'X-Prisme-Referrer': 'not an url',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({})
  })
  expect(response.status).toBe(400)
})

test('non registered domain in Origin header is rejected', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'https://example.com',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      'X-Prisme-Referrer': 'https://example.com/',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({})
  })
  expect(response.status).toBe(400)
})

test('content type different than application/json is rejected', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'https://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      'X-Prisme-Referrer': 'https://mywebsite.localhost/',
      'Content-Type': 'text/plain'
    },
    body: 'abc'
  })
  expect(response.status).toBe(400)
})

test('invalid sessionless custom event', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      // No session associated with this ip.
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({})
  })
  expect(response.status).toBe(400)
})

test('malformed json body', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'https://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      'X-Prisme-Referrer': 'https://mywebsite.localhost/',
      'Content-Type': 'application/json'
    },
    body: '{"foo": "bar and foo", "num": 100' // No closing brace.
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

test('concurrent pageview and custom event', async () => {
  const ipAddr = faker.internet.ip()

  await Promise.all([
    // Custom events first.
    fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
      method: 'POST',
      headers: {
        Origin: 'https://mywebsite.localhost',
        'X-Forwarded-For': ipAddr,
        Referer: 'https://mywebsite.localhost/',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({})
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
  ]).then((results) => results.forEach((resp) => expect(resp.status).toBe(200)))

  const data = await getLatestCustomEvent()

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
      name: 'foo',
      properties: {}
    }
  })
})

test('valid custom event request without body and Content-Type header', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/'
      // 'Content-Type': 'application/json' // not required if no body.
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

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
      name: 'foo',
      properties: {}
    }
  })
})

test('valid custom event with no properties', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({})
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

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
      name: 'foo',
      properties: {}
    }
  })
})

test('valid custom event with JSON bool as body', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(true)
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

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
      name: 'foo',
      properties: {}
    }
  })
})

test('valid custom event with JSON number as body', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(Math.random())
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

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
      name: 'foo',
      properties: {}
    }
  })
})

test('valid custom event with JSON string as body', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(Math.random().toString())
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

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
      timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      domain: 'mywebsite.localhost',
      path: '/',
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      name: 'foo',
      properties: {}
    }
  })
})

test('valid custom event with few properties', async () => {
  const props = {
    x: Math.round(Math.random() * 100),
    y: Math.round(Math.random() * 100)
  }
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(props)
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

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
      name: 'foo',
      properties: props
    }
  })
})

test('valid custom event with lot of properties', async () => {
  const props: Record<number, any> = {}
  for (let i = 0; i < 8; i++) {
    switch (i % 4) {
      case 0: // Bool
        props[i] = Math.random() < 0.5
        break
      case 1: // String
        props[i] = (Math.random() * Number.MAX_SAFE_INTEGER).toString()
        break
      case 2: // Number
        props[i] = Math.random() * Number.MAX_SAFE_INTEGER
        break
      case 3: // Null
        props[i] = null
        break
    }
  }
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(props)
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

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
      name: 'foo',
      properties: props
    }
  })
})

test('valid custom event without X-Prisme-Referrer', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      Referer: 'http://mywebsite.localhost/',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({})
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

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
      name: 'foo',
      properties: {}
    }
  })
})

test('valid custom event without trailing slash in referrer', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      Referer: 'http://mywebsite.localhost',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({})
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

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
      name: 'foo',
      properties: {}
    }
  })
})

test('valid custom event with Windows + Chrome user agent', async () => {
  const userAgent = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.3'

  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost', { userAgent }),
      Referer: 'http://mywebsite.localhost/',
      'Content-Type': 'application/json',
      'User-Agent': userAgent
    },
    body: JSON.stringify({ foo: 'bar' })
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

  expect(data).toMatchObject({
    session: {
      domain: 'mywebsite.localhost',
      entry_path: '/',
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: '/',
      operating_system: 'Windows',
      browser_family: 'Chrome',
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
      name: 'foo',
      properties: { foo: 'bar' }
    }
  })
})

async function getLatestCustomEvent (): Promise<any> {
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

  const customEvents = await client.query({
    query: `SELECT * FROM events_custom WHERE visitor_id = '${session.visitor_id as string}' ORDER BY timestamp DESC LIMIT 1`
  })
  const customEvent = await customEvents.json().then((r: any) => r.data[0])
  if (customEvent === null || customEvent === undefined) return null
  expect(customEvent.visitor_id).toBe(session.visitor_id)
  expect(customEvent.session_uuid).toBe(session.session_uuid)

  customEvent.properties = Object.fromEntries(customEvent.keys.map((key: string, index: number) => [key, JSON.parse(customEvent.values[index])]))
  delete customEvent.keys
  delete customEvent.values

  return { event: customEvent, session }
}
