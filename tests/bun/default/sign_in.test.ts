import { expect, test } from 'bun:test'
import { faker } from '@faker-js/faker'
import { JSDOM } from 'jsdom'

import { postForm } from './utils'
import { signIn, signUp } from './client'
import { PRISME_URL } from '../const'

const seed = new Date().getTime()
console.log('faker seed', seed)
faker.seed(seed)

test('sign in page contains login form', async () => {
  const response = await fetch(PRISME_URL + '/sign_in')
  expect(response.status).toBe(200)
  const html = await response.text()
  const { window: { document } } = new JSDOM(html)

  expect(document.querySelector('form[method=POST]')).not.toBeNil()
  expect(document.querySelector('input[type=email]#email')).not.toBeNil()
  expect(document.querySelector('input[type=password]#password')).not.toBeNil()
})

test('sign in without required email field', async () => {
  const response = await postForm('/sign_in', {
    // email: faker.internet.email(),
    password: faker.internet.password()
  })
  expect(response.status).toBe(400)

  const body = await response.text()
  expect(body).toContain('email invalid: mail: no address')
})

test('sign in without required password field', async () => {
  const response = await postForm('/sign_in', {
    email: faker.internet.email()
    // password: faker.internet.password()
  })
  expect(response.status).toBe(401)

  const body = await response.text()
  expect(body).toContain('login or password incorrect')
})

test('sign in with invalid email (no user registered with that email)', async () => {
  const response = await postForm('/sign_in', {
    email: faker.internet.email(),
    password: faker.internet.password()
  })
  expect(response.status).toBe(401)
})

test('sign in with invalid password', async () => {
  const email = faker.internet.email()
  const password = faker.internet.password()

  await signUp({ name: faker.internet.userName(), email, password })

  const response = await postForm('/sign_in', {
    email,
    password: faker.internet.password()
  })
  expect(response.status).toBe(401)

  const body = await response.text()
  expect(body).toContain('login or password incorrect')
})

test('sign in with valid credentials', async () => {
  const email = faker.internet.email()
  const password = faker.internet.password()

  await signUp({ name: faker.internet.userName(), email, password })

  const cookieHeader = await signIn({
    email,
    password
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
