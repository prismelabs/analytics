import { test, expect } from 'bun:test'

test('X-Forwarded-For is ignored when present', async () => {
  const sourceIp = '10.127.42.1'

  // Generate an access log.
  await fetch('http://prisme.localhost/', {
    headers: {
      'X-Forwarded-For': sourceIp
    }
  })

  // Read access log file.
  const file = Bun.file('/prisme_logs/access.log')
  const text = await file.text()
  const lines = text.split('\n').filter((l) => l !== '')
  const lastLogLine = lines[lines.length - 1]
  const lastLog = JSON.parse(lastLogLine)

  expect(lastLog.source_ip).not.toBe(sourceIp)
})

test('Real IP is used when X-Forwarded-For is missing', async () => {
  // Generate an access log.
  await fetch('http://prisme.localhost/')

  // Read access log file.
  const file = Bun.file('/prisme_logs/access.log')
  const text = await file.text()
  const lines = text.split('\n').filter((l) => l !== '')
  const lastLogLine = lines[lines.length - 1]
  const lastLog = JSON.parse(lastLogLine)

  expect(lastLog.source_ip).toMatch(/[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}/)
})
