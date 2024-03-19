import { test, expect } from 'bun:test'
import { PRISME_PAGEVIEWS_URL } from '../const'
import { faker } from '@faker-js/faker'

const seed = new Date().getTime()
console.log('faker seed', seed)
faker.seed(seed)

test('more than 60 requests per minute are rejected', async () => {
  const clientIp = faker.internet.ip()

  for (let i = 0; i < 100; i++) {
    const response = await fetch(PRISME_PAGEVIEWS_URL, {
      method: 'POST',
      headers: {
        Origin: 'http://mywebsite.localhost',
        'X-Custom-Forwarded-For': clientIp,
        'X-Prisme-Referrer': 'http://mywebsite.localhost'
      }
    })
    if (i < 60) {
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
      Origin: 'http://mywebsite.localhost',
      'X-Custom-Forwarded-For': clientIp,
      'X-Prisme-Referrer': 'http://mywebsite.localhost'
    }
  })
  expect(response.status).toBe(200)
}, { timeout: 120 * 1000 })

test('requests are rate limited based on X-Custom-Forwarded-For header', async () => {
  for (let i = 0; i < 100; i++) {
    const response = await fetch(PRISME_PAGEVIEWS_URL, {
      method: 'POST',
      headers: {
        Origin: 'http://mywebsite.localhost',
        'X-Custom-Forwarded-For': faker.internet.ip(),
        'X-Prisme-Referrer': 'http://mywebsite.localhost'
      }
    })
    expect(response.status).toBe(200)
  }
})
