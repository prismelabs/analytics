import { expect, test } from 'bun:test'
import { faker } from '@faker-js/faker'
import { JSDOM } from 'jsdom'

import { PRISME_URL, postForm } from './utils'
import { signUp } from './client'

const seed = new Date().getTime()
console.log('faker seed', seed)
faker.seed(seed)

test('sign up page contains login form', async () => {
  const response = await fetch(PRISME_URL + '/sign_up')
  expect(response.status).toBe(200)
  const html = await response.text()
  const { window: { document } } = new JSDOM(html)

  expect(document.querySelector('form[method=POST]')).not.toBeNil()
  expect(document.querySelector('input#name')).not.toBeNil()
  expect(document.querySelector('input[type=email]#email')).not.toBeNil()
  expect(document.querySelector('input[type=password]#password')).not.toBeNil()
})

test('sign up without required name field', async () => {
  const response = await postForm('/sign_up', {
    // name: faker.internet.userName(),
    email: faker.internet.email(),
    password: faker.internet.password()
  }
  )
  expect(response.status).toBe(400)

  const body = await response.text()
  expect(body).toContain('user name too short')
})

test('sign up without required email field', async () => {
  const response = await postForm('/sign_up', {
    name: faker.internet.userName(),
    // email: faker.internet.email(),
    password: faker.internet.password()
  }
  )
  expect(response.status).toBe(400)

  const body = await response.text()
  expect(body).toContain('email invalid: mail: no address')
})

test('sign up without required password field', async () => {
  const response = await postForm('/sign_up', {
    name: faker.internet.userName(),
    email: faker.internet.email()
    // password: faker.internet.password()
  }
  )
  expect(response.status).toBe(400)

  const body = await response.text()
  expect(body).toContain('password too short')
})

test('sign up with a valid request', async () => {
  const cookieHeader = await signUp({
    name: faker.internet.userName(),
    email: faker.internet.email(),
    password: faker.internet.password()
  })

  {
    const response = await fetch(PRISME_URL, {
      headers: {
        Cookie: cookieHeader
      },
      redirect: 'error'
    })
    expect(response.status).toBe(200)

    const body = await response.text()
    expect(body).toContain('<title>Home - Prisme Analytics</title>')
  }
})
