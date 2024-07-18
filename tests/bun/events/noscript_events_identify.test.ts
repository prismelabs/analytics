import { expect, test } from 'bun:test'
import { faker } from '@faker-js/faker'

import { createClient } from '@clickhouse/client-web'
import { PRISME_NOSCRIPT_IDENTIFY_EVENTS_URL, PRISME_PAGEVIEWS_URL, PRISME_VISITOR_ID_REGEX, TIMESTAMP_REGEX, UUID_V7_REGEX } from '../const'
import { randomIpWithSession } from '../utils'

const seed = new Date().getTime()
console.log('faker seed', seed)
faker.seed(seed)

test('POST request instead of GET request', async () => {
  const response = await fetch(PRISME_NOSCRIPT_IDENTIFY_EVENTS_URL, {
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
  const response = await fetch(PRISME_NOSCRIPT_IDENTIFY_EVENTS_URL, {
    method: 'GET',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      Referer: 'not an url',
      'Content-Type': 'application/json'
    }
  })
  expect(response.status).toBe(400)
})

test('non registered domain in Origin header is rejected', async () => {
  const response = await fetch(PRISME_NOSCRIPT_IDENTIFY_EVENTS_URL, {
    method: 'GET',
    headers: {
      Origin: 'https://example.com',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost'),
      Referer: 'https://example.com/foo?bar=baz#qux'
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

test('concurrent pageview and identify events', async () => {
  const ipAddr = faker.internet.ip()
  const visitorId = `visitor-id-${Math.random()}`

  await Promise.all([
    // Identify events first.
    fetch(PRISME_NOSCRIPT_IDENTIFY_EVENTS_URL + `?visitorId=${visitorId}`, {
      method: 'GET',
      headers: {
        Origin: 'https://mywebsite.localhost',
        'X-Forwarded-For': ipAddr,
        Referer: 'https://mywebsite.localhost/foo'
      }
    }),
    // Pageview concurrently.
    // This pageview will create session that will be used for both events.
    fetch(PRISME_PAGEVIEWS_URL, {
      method: 'POST',
      headers: {
        Origin: 'https://mywebsite.localhost',
        'X-Forwarded-For': ipAddr,
        Referer: 'https://mywebsite.localhost/foo'
      }
    })
  ]).then((results) => results.forEach((resp) => expect(resp.status).toBe(200)))

  // Session contains visitor ID A.
  const session = await getLatestSession()
  const sessionUuid = session.session_uuid
  expect(session).toMatchObject({
    visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
    session_uuid: expect.stringMatching(UUID_V7_REGEX),
    version: 1
  })

  // Check that user exists.
  const user = await getLatestUser()
  expect(user).toMatchObject({
    // Visitor ID B is used to store user props.
    visitor_id: visitorId,
    initial_session_uuid: sessionUuid,
    initial_session_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    latest_session_uuid: sessionUuid,
    latest_session_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    initialProperties: {},
    properties: {}
  })
})

test('valid identify with visitor_id only', async () => {
  const visitorIdA = `visitor-id-${Math.random()}`
  const visitorIdB = `visitor-id-${Math.random()}`
  const ipAddr = await randomIpWithSession('mywebsite.localhost', { visitorId: visitorIdA })
  const response = await fetch(PRISME_NOSCRIPT_IDENTIFY_EVENTS_URL + `?visitorId=${visitorIdB}`, {
    method: 'GET',
    headers: {
      Origin: 'https://mywebsite.localhost',
      'X-Forwarded-For': ipAddr,
      Referer: 'https://mywebsite.localhost/foo'
    }
  })
  expect(response.status).toBe(200)

  const user = await getLatestUser()
  expect(user).toMatchObject({
    // Visitor ID B is used to store user props.
    visitor_id: visitorIdB,
    initial_session_uuid: expect.stringMatching(UUID_V7_REGEX),
    initial_session_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    latest_session_uuid: expect.stringMatching(UUID_V7_REGEX),
    latest_session_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    initialProperties: {},
    properties: {}
  })

  // Session contains visitor ID A.
  let session = await getLatestSession()
  expect(session).toMatchObject({
    // visitor ID A and not B as identify event doesn't change visitor id of
    // previous pageviews.
    visitor_id: visitorIdA,
    session_uuid: expect.stringMatching(UUID_V7_REGEX),
    version: 1
  })

  // View another page.
  {
    const response = await fetch(PRISME_PAGEVIEWS_URL, {
      method: 'POST',
      headers: {
        Origin: 'https://mywebsite.localhost',
        'X-Forwarded-For': ipAddr,
        Referer: 'https://mywebsite.localhost/bar',
        'X-Prisme-Document-Referrer': 'https://mywebsite.localhost/foo'
      }
    })
    expect(response.status).toBe(200)
  }

  // Session contains visitorIdB now.
  session = await getLatestSession()
  expect(session).toMatchObject({
    visitor_id: visitorIdB,
    session_uuid: expect.stringMatching(UUID_V7_REGEX),
    version: 2
  })
})

test('multiple identify events for same visitor id with different "set" props overwrite previous identify props', async () => {
  const visitorId = `visitor-id-${Math.random()}`
  let date = new Date().toUTCString()
  let response = await fetch(PRISME_NOSCRIPT_IDENTIFY_EVENTS_URL + `?visitorId=${visitorId}&${propsToQuery({
        date,
        foo: 'bar',
        bar: undefined,
        baz: 1,
        nested: {
          foo: 'bar2'
        },
        bool: true
      }, 'set-')}`, {
    method: 'GET',
    headers: {
      Origin: 'https://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost', { visitorId }),
      Referer: 'https://mywebsite.localhost/foo'
    }
  })
  expect(response.status).toBe(200)

  let user = await getLatestUser()
  expect(user).toMatchObject({
    visitor_id: visitorId,
    initial_session_uuid: expect.stringMatching(UUID_V7_REGEX),
    initial_session_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    latest_session_uuid: expect.stringMatching(UUID_V7_REGEX),
    latest_session_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    initialProperties: {},
    properties: {
      date,
      foo: 'bar',
      baz: 1,
      nested: {
        foo: 'bar2'
      },
      bool: true
    }
  })

  // Second identify event.
  date = new Date().toUTCString() // Update date.
  response = await fetch(PRISME_NOSCRIPT_IDENTIFY_EVENTS_URL + `?visitorId=${visitorId}&${propsToQuery({
        date,
        foo: 'bar',
        bar: undefined,
        baz: 2,
        nested: {
          foo: 'bar2'
        },
        bool: true
      }, 'set-')}`, {
    method: 'GET',
    headers: {
      Origin: 'https://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost', { visitorId }),
      Referer: 'https://mywebsite.localhost/foo'
    }
  })
  expect(response.status).toBe(200)

  user = await getLatestUser()
  expect(user).toMatchObject({
    visitor_id: visitorId,
    initial_session_uuid: expect.stringMatching(UUID_V7_REGEX),
    initial_session_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    latest_session_uuid: expect.stringMatching(UUID_V7_REGEX),
    latest_session_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    initialProperties: {},
    properties: {
      date,
      foo: 'bar',
      baz: 2,
      nested: {
        foo: 'bar2'
      },
      bool: true
    }
  })
})

test('multiple identify events for same visitor id with different "setOnce" props do not overwrite props', async () => {
  const visitorId = `visitor-id-${Math.random()}`
  const date = new Date().toUTCString()
  let response = await fetch(PRISME_NOSCRIPT_IDENTIFY_EVENTS_URL + `?visitorId=${visitorId}&${propsToQuery({
        date,
        foo: 'bar',
        bar: undefined,
        baz: 1,
        nested: {
          foo: 'bar2'
        },
        bool: true
      }, 'set-once-')}`, {
    method: 'GET',
    headers: {
      Origin: 'https://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost', { visitorId }),
      Referer: 'https://mywebsite.localhost/foo'
    }
  })
  expect(response.status).toBe(200)

  let user = await getLatestUser()
  expect(user).toMatchObject({
    visitor_id: visitorId,
    latest_session_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    initial_session_uuid: expect.stringMatching(UUID_V7_REGEX),
    latest_session_uuid: expect.stringMatching(UUID_V7_REGEX),
    initialProperties: {
      date,
      foo: 'bar',
      baz: 1,
      nested: {
        foo: 'bar2'
      },
      bool: true
    },
    properties: { }
  })

  // Second identify event.
  response = await fetch(PRISME_NOSCRIPT_IDENTIFY_EVENTS_URL + `?visitorId=${visitorId}&${propsToQuery({
        date2: date,
        date,
        foo: 'bar',
        bar: undefined,
        baz: 2,
        nested: {
          foo: 'bar2'
        },
        bool: true
      }, 'set-once-')}`, {
    method: 'GET',
    headers: {
      Origin: 'https://mywebsite.localhost',
      'X-Forwarded-For': await randomIpWithSession('mywebsite.localhost', { visitorId }),
      Referer: 'https://mywebsite.localhost/foo'
    }
  })
  expect(response.status).toBe(200)

  user = await getLatestUser()
  expect(user).toMatchObject({
    visitor_id: visitorId,
    latest_session_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    initial_session_uuid: expect.stringMatching(UUID_V7_REGEX),
    latest_session_uuid: expect.stringMatching(UUID_V7_REGEX),
    initialProperties: {
      // TODO: date2 should be added.
      // date2: date, // Not added.
      date, // Unchanged.
      foo: 'bar',
      baz: 1, // Unchanged.
      nested: {
        foo: 'bar2'
      },
      bool: true
    },
    properties: {}
  })
})

async function getLatestUser (): Promise<any> {
  // Wait for clickhouse to ingest batch.
  Bun.sleepSync(1000)

  const client = createClient({
    host: 'http://clickhouse.localhost:8123',
    username: 'clickhouse',
    password: 'password',
    database: 'prisme'
  })

  const user = await client.query({
    query: 'SELECT * FROM users_props ORDER BY latest_session_timestamp DESC LIMIT 1'
  }).then(props => props.json())
    .then((r: any) => r.data[0])

  user.initialProperties = Object.fromEntries(
    user.initial_keys?.map((key: string, index: number) =>
      [key, JSON.parse(user.initial_values[index])])
  )
  delete user.initial_keys
  delete user.initial_values

  user.properties = Object.fromEntries(
    user.keys?.map((key: string, index: number) =>
      [key, JSON.parse(user.values[index])])
  )
  delete user.keys
  delete user.values

  return user
}

async function getLatestSession (): Promise<any> {
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

  return session
}

function propsToQuery (props: Record<string, any>, prefix?: string): string {
  return Object.entries(props).filter(([k, v]) => k !== undefined && v !== undefined)
    .map(([k, v]) => (prefix ?? '') + k + '=' + JSON.stringify(v)).join('&')
}
