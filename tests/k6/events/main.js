import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
	discardResponseBodies: true,
	scenarios: {
		sharedIterationsPageViewEvents: {
			executor: "shared-iterations",
			vus: 4096,
			iterations: 2 ** 19,
			exec: "pageView",
		},
		sharedIterationsCustomEvents: {
			executor: "shared-iterations",
			vus: 4096,
			iterations: 2 ** 19,
			exec: "customEvent",
		}
	}
}

const origins = [
	"mywebsite.localhost",
	"foo.mywebsite.localhost",
	"someoneelsewebsite.com"
]

export function pageView() {
	const origin = [
				randomItem(["http", "https"]),
				"://",
				randomItem(origins),
	].join('')
	const docReferrer = randomItem([
		undefined,
		"https://google.com",
		"https://duckduckgo.com",
		"https://qwant.com",
		"https://github.com",
		origin,
	])

	const res = http.post('http://prisme.localhost/api/v1/events/pageviews', null, {
		headers: {
			"Origin": origin,
			"X-Prisme-Referrer": [
				origin,
				randomItem(["/", "/foo", "/bar", "qux", "/foo/"])
			].join(''),
			"X-Forwarded-For": randomIP(),
			"X-Prisme-Document-Referrer": docReferrer
		}
	})
}

export function customEvent() {
	const origin = [
				randomItem(["http", "https"]),
				"://",
				randomItem(origins),
	].join('')
	const docReferrer = randomItem([
		undefined,
		"https://google.com",
		"https://duckduckgo.com",
		"https://qwant.com",
		"https://github.com",
		...origins,
	])

	const res = http.post(`http://prisme.localhost/api/v1/events/custom/${"foo"}`, JSON.stringify({x: 1024, y: 4096}), {
		headers: {
			"Content-Type": "application/json",
			"Origin": origin,
			"X-Prisme-Referrer": [
				origin,
				randomItem(["/", "/foo", "/bar", "qux", "/foo/"])
			].join(''),
			"X-Forwarded-For": randomIP(),
			"X-Prisme-Document-Referrer": docReferrer
		}
	})
}

function randomItem(items) {
	const index = Math.floor(Math.random() * items.length)
	return items[index]
}

function randomIP() {
	const addr = [10, 10]
	// for (let i = 0; i < 2; i++) {
	addr.push(Math.floor(Math.random() * 16))
	addr.push(Math.floor(Math.random() * 255))
	// }

	return addr.map((b) => b.toString()).join('.')
}
