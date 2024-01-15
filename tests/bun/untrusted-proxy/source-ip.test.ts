import { test, expect } from 'bun:test'

const UUID_V4_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i

test('X-Request-Id is ignored when present', async () => {
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
