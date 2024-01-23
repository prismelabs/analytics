import { test, expect } from 'bun:test'
import { UUID_V4_REGEX } from '../const'

test('X-Request-Id is used when present', async () => {
  // Generate an access log.
  await fetch('http://prisme.localhost/', {
    headers: {
      'X-Request-Id': 'foo'
    }
  })

  // Read access log file.
  const file = Bun.file('/prisme_logs/access.log')
  const text = await file.text()
  const lines = text.split('\n').filter((l) => l !== '')
  const lastLogLine = lines[lines.length - 1]
  const lastLog = JSON.parse(lastLogLine)

  expect(lastLog.request_id).toMatch(UUID_V4_REGEX)
})

test('Random UUID v4 is used when X-Request-Id is missing', async () => {
  // Generate an access log.
  await fetch('http://prisme.localhost/')

  // Read access log file.
  const file = Bun.file('/prisme_logs/access.log')
  const text = await file.text()
  const lines = text.split('\n').filter((l) => l !== '')
  const lastLogLine = lines[lines.length - 1]
  const lastLog = JSON.parse(lastLogLine)

  expect(lastLog.request_id).toMatch(UUID_V4_REGEX)
})
