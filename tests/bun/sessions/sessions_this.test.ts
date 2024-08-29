import { expect, test } from 'bun:test'
import { faker } from '@faker-js/faker'
import { PRISME_PAGEVIEWS_URL, PRISME_SESSIONS_THIS_URL, PRISME_VISITOR_ID_REGEX, UUID_V7_REGEX } from '../const'

const seed = new Date().getTime()
console.log('faker seed', seed)
faker.seed(seed)

test('', async () => {
  const ipAddr = '8.8.8.8'
  const userAgent = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.3'
  // Create session.
  let response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': ipAddr,
      Referer: 'http://foo.mywebsite.localhost/?utm_source=my_utm_source',
      'X-Prisme-Document-Referrer': 'http://example.com/',
      'User-Agent': userAgent
    }
  })
  expect(response.status).toBe(200)

  // Retrieve session.
  response = await fetch(PRISME_SESSIONS_THIS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': ipAddr,
      Referer: 'http://foo.mywebsite.localhost/',
      'User-Agent': userAgent
    }
  })
  expect(response.status).toBe(200)

  let data = await response.json()
  const visitorId = data.VisitorId
  const sessionUuid = data.SessionUuid
  expect(data).toMatchObject({
    Client: {
      OperatingSystem: 'Windows',
      BrowserFamily: 'Chrome',
      Device: 'Other',
      IsBot: false
    },
    CountryCode: 'US',
    PageUri: 'http://foo.mywebsite.localhost/?utm_source=my_utm_source',
    PageviewCount: 1,
    ReferrerUri: 'http://example.com/',
    Utm: {
      Campaign: '',
      Content: '',
      Medium: '',
      Source: 'my_utm_source',
      Term: ''
    },
    VisitorId: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
    SessionUuid: expect.stringMatching(UUID_V7_REGEX)
  })

  // Second page view.
  response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': ipAddr,
      'X-Prisme-Document-Referrer': 'http://foo.mywebsite.localhost/',
      Referer: 'http://foo.mywebsite.localhost/bar',
      'User-Agent': userAgent
    }
  })
  expect(response.status).toBe(200)

  // Retrieve session.
  response = await fetch(PRISME_SESSIONS_THIS_URL, {
    method: 'POST',
    headers: {
      Origin: 'http://foo.mywebsite.localhost',
      'X-Forwarded-For': ipAddr,
      Referer: 'http://foo.mywebsite.localhost',
      'User-Agent': userAgent
    }
  })
  expect(response.status).toBe(200)

  data = await response.json()
  expect(data).toMatchObject({
    Client: {
      OperatingSystem: 'Windows',
      BrowserFamily: 'Chrome',
      Device: 'Other',
      IsBot: false
    },
    CountryCode: 'US',
    PageUri: 'http://foo.mywebsite.localhost/?utm_source=my_utm_source',
    PageviewCount: 2,
    ReferrerUri: 'http://example.com/',
    Utm: {
      Campaign: '',
      Content: '',
      Medium: '',
      Source: 'my_utm_source',
      Term: ''
    },
    VisitorId: visitorId,
    SessionUuid: sessionUuid
  })
})
