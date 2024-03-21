import { expect, test } from 'bun:test'
import { faker } from '@faker-js/faker'

import { createClient } from '@clickhouse/client-web'
import { PRISME_CUSTOM_EVENTS_URL, TIMESTAMP_REGEX } from '../const'

const seed = new Date().getTime()
console.log('faker seed', seed)
faker.seed(seed)

test('GET request instead of POST request', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'GET',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/foo',
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
      'X-Forwarded-For': faker.internet.ip(),
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
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'https://example.com/foo?bar=baz#qux',
      'Content-Type': 'application/json',
      body: JSON.stringify({})
    }
  })
  expect(response.status).toBe(400)
})

test('content type different than application/json is rejected', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'https://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'https://mywebsite.localhost/foo?bar=baz#qux',
      'Content-Type': 'text/plain'
    }
  })
  expect(response.status).toBe(400)
})

test('valid custom event request without body', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/index.html',
      'Content-Type': 'application/json'
    }
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: 'mywebsite.localhost',
    path: '/index.html',
    name: 'foo',
    properties: {}
  })
})

test('valid custom event with no properties', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/index.html',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({})
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: 'mywebsite.localhost',
    path: '/index.html',
    name: 'foo',
    properties: {}
  })
})

test('valid custom event with JSON bool as body', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/index.html',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(true)
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: 'mywebsite.localhost',
    path: '/index.html',
    name: 'foo',
    properties: {}
  })
})

test('valid custom event with JSON number as body', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/index.html',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(Math.random())
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: 'mywebsite.localhost',
    path: '/index.html',
    name: 'foo',
    properties: {}
  })
})

test('valid custom event with JSON string as body', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/index.html',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(Math.random().toString())
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: 'mywebsite.localhost',
    path: '/index.html',
    name: 'foo',
    properties: {}
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
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/index.html',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(props)
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: 'mywebsite.localhost',
    path: '/index.html',
    name: 'foo',
    properties: props
  })
})

test('valid custom event with lot of properties', async () => {
  const props: Record<number, number> = {}
  for (let i = 0; i < 1024; i++) {
    props[i] = Math.round(Math.random() * 100)
  }
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/index.html',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(props)
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: 'mywebsite.localhost',
    path: '/index.html',
    name: 'foo',
    properties: props
  })
})

test('valid custom event without X-Prisme-Referrer', async () => {
  const response = await fetch(PRISME_CUSTOM_EVENTS_URL + '/foo', {
    method: 'POST',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      Referer: 'http://mywebsite.localhost/index.html',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({})
  })
  expect(response.status).toBe(200)

  const data = await getLatestCustomEvent()

  expect(data).toMatchObject({
    timestamp: expect.stringMatching(TIMESTAMP_REGEX),
    domain: 'mywebsite.localhost',
    path: '/index.html',
    name: 'foo',
    properties: {}
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

  const rows = await client.query({
    query: 'SELECT * FROM prisme.events_custom ORDER BY timestamp DESC LIMIT 1;'
  })
  const row = await rows.json().then((r: any) => r.data[0])
  row.properties = Object.fromEntries(row.keys.map((key: string, index: number) => [key, JSON.parse(row.values[index])]))
  delete row.keys
  delete row.values
  return row
}
