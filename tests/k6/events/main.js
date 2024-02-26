import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
	discardResponseBodies: true,
	scenarios: {
		sharedIterations: {
			executor: "shared-iterations",
			vus: 4096,
			iterations: 2 ** 17,
		}
	}
}

export default function() {
	const res = http.post('http://prisme.localhost/api/v1/events/pageviews', null, {
		headers: {
			"X-Prisme-Referrer": [
				randomItem(["http", "https"]),
				"://",
				randomItem(["mywebsite.localhost", "foo.mywebsite.localhost", "someoneelsewebsite.com"]),
				randomItem(["/", "/foo", "/bar", "qux"])
			].join(''),
			"X-Prisme-Document-Referrer": randomItem([undefined, "https://google.com", "https://duckduckgo.com", "https://qwant.com", "https://github.com"]),
			"X-Forwarded-For": randomIP()
		}
	})

	sleep(0.5)
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
