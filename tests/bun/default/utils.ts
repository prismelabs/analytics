import { PRISME_URL } from '../const'

export async function postForm (path: string, body: Record<string, any>): Promise<ReturnType<typeof fetch>> {
  const response = await fetch(`${PRISME_URL}${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    body: Object.entries(body).map(([key, value]) =>
      encodeURIComponent(key) + '=' + encodeURIComponent(value))
      .join('&'),
    redirect: 'manual'
  })

  return response
}

export interface Cookie {
  name: string
  value: string
  attributes: {
    'max-age'?: number
    path?: string
    HttpOnly?: boolean
    secure?: boolean
    SameSite?: 'Strict' | 'Lax'
  }
}

export function parseSetCookie (rawCookie: string): Cookie {
  const [[name, value], ...attrs] = rawCookie.split(';').map((pair) => pair.trim().split('='))
  const cookie: Partial<Cookie> = {
    name,
    value,
    attributes: {}
  }

  for (const [name, value] of attrs) {
    switch (name.toLowerCase()) {
      case 'max-age':
        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        cookie.attributes!['max-age'] = Number.parseInt(value)
        break

      case 'path':
        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        cookie.attributes!.path = value
        break

      case 'httponly':
        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        cookie.attributes!.HttpOnly = true
        break

      case 'secure':
        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        cookie.attributes!.secure = true
        break

      case 'samesite':
        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        cookie.attributes!.SameSite = value as Cookie['attributes']['SameSite']
        break

      default:
        throw new Error('unknown cookie attribute ' + name)
    }
  }

  return cookie as Cookie
}
