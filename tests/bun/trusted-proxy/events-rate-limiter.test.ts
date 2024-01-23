import { test, expect } from 'bun:test'
import { PRISME_PAGEVIEWS_URL } from '../const'
import { faker } from '@faker-js/faker'

const seed = new Date().getTime()
console.log('faker seed', seed)
faker.seed(seed)

test('more than 10 requests per minute are rejected', async () => {
  const clientIp = faker.internet.ip()

  for (let i = 0; i < 20; i++) {
    const response = await fetch(PRISME_PAGEVIEWS_URL, {
      method: 'POST',
      headers: {
        'X-Forwarded-For': clientIp,
        'X-Prisme-Referrer': 'http://mywebsite.localhost'
      }
    })
    if (i < 10) {
      expect(response.status).toBe(200)
    } else {
      expect(response.status).toBe(429)
    }
  }

  // Wait a minute.
  Bun.sleepSync(60 * 1000)

  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'POST',
    headers: {
      'X-Forwarded-For': clientIp,
      'X-Prisme-Referrer': 'http://mywebsite.localhost'
    }
  })
  expect(response.status).toBe(200)
}, { timeout: 120 * 1000 })

test('requests are rate limited based on X-Forwarded-For header', async () => {
  for (let i = 0; i < 50; i++) {
    const response = await fetch(PRISME_PAGEVIEWS_URL, {
      method: 'POST',
      headers: {
        'X-Forwarded-For': faker.internet.ip(),
        'X-Prisme-Referrer': 'http://mywebsite.localhost'
      }
    })
    expect(response.status).toBe(200)
  }
})
