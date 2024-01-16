import { expect } from 'bun:test'

import { parseSetCookie, postForm } from './utils'
import { UUID_V4_REGEX } from './matchers'

export interface SignUpRequest {
  name: string
  email: string
  password: string
}

export type SessionCookie = string

export async function signUp (body: SignUpRequest): Promise<SessionCookie> {
  const response = await postForm('/sign_up', body)
  expect(response.status).toBe(302)
  expect(response.headers.getSetCookie()).toHaveLength(1)

  const cookie = parseSetCookie(response.headers.getSetCookie()[0])
  expect({ ...cookie }).toMatchObject({
    name: 'prisme_session_id',
    value: expect.stringMatching(UUID_V4_REGEX),
    attributes: {
      'max-age': 86400,
      path: '/',
      HttpOnly: true,
      secure: true,
      SameSite: 'Strict'
    }
  })
  expect(response.headers.get('Location')).toBe('/')

  return `${cookie.name}=${cookie.value}`
}
