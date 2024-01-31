import { test, expect } from 'bun:test'

test('X-Custom-Forwarded-For is used when present', async () => {
  const sourceIp = '10.127.42.1'

  // Generate an access log.
  await fetch('http://prisme.localhost/', {
    headers: {
      'X-Custom-Forwarded-For': sourceIp
    }
  })

  // Read access log file.
  const file = Bun.file('/prisme_logs/access.log')
  const text = await file.text()
  const lines = text.split('\n').filter((l) => l !== '')
  const lastLogLine = lines[lines.length - 1]
  const lastLog = JSON.parse(lastLogLine)

  expect(lastLog.source_ip).toBe(sourceIp)
})

test('Real IP is used when X-Custom-Forwarded-For is missing', async () => {
  // Generate an access log.
  await fetch('http://prisme.localhost/')

  // Read access log file.
  const file = Bun.file('/prisme_logs/access.log')
  const text = await file.text()
  const lines = text.split('\n').filter((l) => l !== '')
  const lastLogLine = lines[lines.length - 1]
  const lastLog = JSON.parse(lastLogLine)

  expect(lastLog.source_ip).toBe('')
})
