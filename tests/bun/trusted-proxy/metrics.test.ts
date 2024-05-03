import { test, expect } from 'bun:test'
import { PRISME_METRICS_URL, PRISME_PAGEVIEWS_URL } from '../const'
import { faker } from '@faker-js/faker'

test('http total request metric', async () => {
  // Fetch metrics before starting test.
  const beforeMetrics = await fetchAndParseMetrics()

  // Invalid event request.
  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: 'GET',
    headers: {
      Origin: 'http://mywebsite.localhost',
      'X-Forwarded-For': faker.internet.ip(),
      'X-Prisme-Referrer': 'http://mywebsite.localhost/foo'
    }
  })
  expect(response.status).toBe(405)

  // Retrieve metrics.
  const afterMetrics = await fetchAndParseMetrics()

  const labels = {
    method: 'GET',
    path: '/api/v1/events/*',
    status: '405'
  }
  const expectedValue = Number.parseFloat(counterValue(beforeMetrics, 'http_requests_total', labels)) + 1

  expect(counterValue(afterMetrics, 'http_requests_total', labels)).toBe(expectedValue.toString())
  expect(histogramCountValue(afterMetrics, 'http_requests_duration_seconds', labels)).toBe(expectedValue.toString())
})

interface Metrics {
  counter: Record<string, Metric[]>
  summary: Record<string, Metric[]>
  gauge: Record<string, Metric[]>
  histogram: Record<string, Metric[]>
}

interface Metric {
  labels: Record<string, string>
  value: string
}

async function fetchMetrics (): Promise<string> {
  const resp = await fetch(PRISME_METRICS_URL)
  expect(resp.ok).toBeTrue()

  return await resp.text()
}

async function parseMetrics (prometheusMetrics: string): Promise<Metrics> {
  const metrics: Metrics = {
    counter: {},
    summary: {},
    gauge: {},
    histogram: {}
  }

  const lines = prometheusMetrics.split('\n')
  let metricType: keyof Metrics | undefined
  for (const line of lines) {
    const words = line.split(' ')

    if (line.startsWith('# HELP') || line === '') continue
    if (line.startsWith('# TYPE')) {
      metricType = words[words.length - 1] as keyof Metrics
      continue
    }

    // Check type was defined before.
    if (metricType === undefined) {
      throw new Error('invalid prometheus metric text')
    }

    const [metricNameAndLabels, metricValue] = words
    const [metricName, rawLabels] = metricNameAndLabels.split(/(\{|\})/)
      .filter((s) => s !== '' && s !== '{' && s !== '}')
    let labels = {}
    if (rawLabels?.length > 0) {
      labels = Object.fromEntries(rawLabels.split(',').map((label) => label.split('='))
        .map(([key, value]) => [key, value.slice(1, -1)]))
    }

    metrics[metricType][metricName] = metrics[metricType][metricName] ?? []
    metrics[metricType][metricName].push({ labels, value: metricValue })
  }

  return metrics
}

async function fetchAndParseMetrics (): Promise<Metrics> {
  return await parseMetrics(await fetchMetrics())
}

function counterValue (metrics: Metrics, name: string, labels: Record<string, string>): string {
  for (const { labels: l, value } of metrics.counter[name]) {
    if (jsonEq(l, labels)) {
      return value
    }
  }

  return '0'
}

function histogramCountValue (metrics: Metrics, name: string, labels: Record<string, string>): string {
  for (const { labels: l, value } of metrics.histogram[name + '_count']) {
    if (jsonEq(l, labels)) {
      return value
    }
  }

  return '0'
}

function jsonEq (a: Record<string, any>, b: Record<string, any>): boolean {
  const aKeys = Object.keys(a).sort()
  const bKeys = Object.keys(b).sort()

  if (aKeys.length !== bKeys.length) return false

  // Sort keys of objects so JSON serialization can be compared.
  const sortedA: Record<string, any> = {}
  const sortedB: Record<string, any> = {}
  for (const key of aKeys) {
    sortedA[key] = a[key]
    sortedB[key] = b[key]
  }

  return JSON.stringify(sortedA) === JSON.stringify(sortedB)
}
