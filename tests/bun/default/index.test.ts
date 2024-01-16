import { expect, test } from 'bun:test'
import { PRISME_URL } from './utils'

test('anonymous request is redirected to /sign_in', async () => {
  const response = await fetch(PRISME_URL, {
    redirect: 'manual'
  })
  expect(response.status).toBe(302)
  expect(response.headers.get('Location')).toBe('/sign_in')
})

test('request with invalid session id is redirected to /sign_in', async () => {
  const response = await fetch(PRISME_URL, {
    headers: {
      Cookie: 'prisme_session_id=1234'
    },
    redirect: 'manual'
  })
  expect(response.status).toBe(302)
  expect(response.headers.get('Location')).toBe('/sign_in')
})
