import { expect } from "@std/expect";
import { faker } from "@faker-js/faker";

import { PRISME_PAGEVIEWS_URL } from "../const.ts";
import { sleep } from "../utils.ts";

const seed = new Date().getTime();
console.log("faker seed", seed);
faker.seed(seed);

Deno.test("more than 60 requests per minute are rejected", async () => {
  const clientIp = faker.internet.ip();

  for (let i = 0; i < 100; i++) {
    const response = await fetch(PRISME_PAGEVIEWS_URL, {
      method: "POST",
      headers: {
        Origin: "http://mywebsite.localhost",
        "X-Custom-Forwarded-For": clientIp,
        "X-Prisme-Referrer": "http://mywebsite.localhost",
      },
    });
    if (i < 60) {
      expect(response.status).toBe(200);
    } else {
      expect(response.status).toBe(429);
    }
  }

  // Wait a minute.
  await sleep(60 * 1000);

  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: "POST",
    headers: {
      Origin: "http://mywebsite.localhost",
      "X-Custom-Forwarded-For": clientIp,
      "X-Prisme-Referrer": "http://mywebsite.localhost",
    },
  });
  expect(response.status).toBe(200);
});

Deno.test("requests are rate limited based on X-Custom-Forwarded-For header", async () => {
  for (let i = 0; i < 100; i++) {
    const response = await fetch(PRISME_PAGEVIEWS_URL, {
      method: "POST",
      headers: {
        Origin: "http://mywebsite.localhost",
        "X-Custom-Forwarded-For": faker.internet.ip(),
        "X-Prisme-Referrer": "http://mywebsite.localhost",
      },
    });
    expect(response.status).toBe(200);
  }
});
