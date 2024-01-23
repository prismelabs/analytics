import { test, expect } from 'bun:test'
import { PRISME_PAGEVIEWS_URL } from '../const'
import { faker } from '@faker-js/faker'

const seed = new Date().getTime()
console.log('faker seed', seed)
faker.seed(seed)

test('requests are rate limited based on X-Forwarded-For header', async () => {
  for (let i = 0; i < 50; i++) {
    const response = await fetch(PRISME_PAGEVIEWS_URL, {
      method: 'POST',
      headers: {
        'X-Forwarded-For': faker.internet.ip(), // ignored.
        'X-Prisme-Referrer': 'http://mywebsite.localhost'
      }
    })
    if (i < 10) {
      expect(response.status).toBe(200)
    } else {
      expect(response.status).toBe(429)
    }
  }
})
