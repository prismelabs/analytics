import http from 'k6/http';
import { Counter } from 'k6/metrics';
import { sleep } from 'k6';

const directTrafficRate = 0.1;
const customEventRate = 0.3;
const errorRate = 0.1;
const bounceRate = 0.5;
const exitRate = 0.1;

export const options = {
	thresholds: {
		// Thresholds so tags appear in CLI report.
		'http_reqs{event_type:pageview}': ['count >= 0'],
		'http_reqs{event_type:custom}': ['count >= 0'],
	},
	discardResponseBodies: true,
	scenarios: {
		sharedIterationsPageViewEvents: {
			executor: "shared-iterations",
			vus: 4096,
			iterations: 2 ** 17,
		},
	}
}

const origins = [
	"mywebsite.localhost",
	"foo.mywebsite.localhost",
]

export default function () {
	const origin = [
		randomItem(["http", "https"]),
		"://",
		randomItem(origins),
	].join('')

	const ipAddr = randomIP()

	// Entry pageview.
	const response = pageView({entryPageView: true, origin, ipAddr})
	if (response.status !== 200) {
		// console.error("entry pageview", response.status_text, response.error)
		return
	}

	// Custom events.
	while (Math.random() < customEventRate) {
		const response = customEvent({origin, ipAddr})
		if (response.status !== 200) {
			// console.error("custom event", response.status_text, response.error)
			return
		}
	}

	// Bounces.
	if (Math.random() < bounceRate) {
		return
	}

	let events = 0
	while (events < 30) {
		// Pageview.
		const response = pageView({origin, ipAddr})
		if (response.status !== 200) {
			// console.error("pageview", response.status_text, response.error)
			return
		}
		events++

		// Custom events.
		while (Math.random() < customEventRate) {
			const response = customEvent({origin, ipAddr})
			if (response.status !== 200) {
				// console.error("custom event", response.status_text, response.error)
				return
			}
			events++
		}

		// Exit rate.
		if (Math.random() < exitRate) {
			return
		}
	}
}

function pageView({ entryPageView = false, origin, ipAddr }) {
	const headers = {
		"Origin": origin,
		"X-Prisme-Referrer": [
			origin,
			randomItem(["/", "/foo", "/bar", "qux", "/foo/"])
		].join(''),
		"X-Prisme-Document-Referrer": origin,
		"X-Forwarded-For": ipAddr,
	}

	if (entryPageView) {
		if (Math.random() < directTrafficRate) {
			delete headers["X-Prisme-Document-Referrer"]
		} else {
			headers["X-Prisme-Document-Referrer"] = randomItem([
				"https://google.com",
				"https://duckduckgo.com",
				"https://qwant.com",
				"https://github.com",
			])
		}
	}

	if (Math.random() < errorRate) {
		// Invalid origin.
		headers["Origin"] = "undefined"
	}

	const response = http.post(
		'http://prisme.localhost/api/v1/events/pageviews',
		null,
		{ headers, tags: { event_type: "pageview" } }
	)

	return response
}

function customEvent({origin, ipAddr}) {
	const headers = {
		"Origin": origin,
		"Content-Type": "application/json",
		"X-Prisme-Referrer": [
			origin,
			randomItem(["/", "/foo", "/bar", "qux", "/foo/"])
		].join(''),
		"X-Forwarded-For": ipAddr,
	}

	if (Math.random() < errorRate) {
		// Invalid origin.
		headers["Origin"] = "undefined"
	}

	const eventName = randomItem(["click", "empty", "big", "download"])
	let body = {}
	switch (eventName) {
		case "click":
			body = {x: Math.round(Math.random() * 100), y: Math.round(Math.random() * 100) }
			break;

		case "empty":
			break;

		case "big":
			for (let i = 0; i < 32; i++) {
				body[i] = i
			}
			break;

		case "download":
			body["file"] = randomItem(["file.pdf", "summary.pdf", "company.pdf"])
			break;

		default:
			throw new Error("unknown event name: " + eventName)
	}

	const response = http.post(
		'http://prisme.localhost/api/v1/events/custom/' + eventName,
		JSON.stringify(body),
		{ headers, tags: { event_type: "custom" } }
	)

	return response
}

function randomItem(items) {
	const index = Math.floor(Math.random() * items.length)
	return items[index]
}

function randomIP() {
	const addr = []
	for (let i = 0; i < 4; i++) {
		addr.push(Math.floor(Math.random() * 255))
	}

	return addr.map((b) => b.toString()).join('.')
}
