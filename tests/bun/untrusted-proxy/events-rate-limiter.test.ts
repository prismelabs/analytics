import { expect } from "@std/expect";
import { faker } from "@faker-js/faker";

import { PRISME_PAGEVIEWS_URL } from "../const.ts";

const seed = new Date().getTime();
console.log("faker seed", seed);
faker.seed(seed);

Deno.test("requests are rate limited based on X-Forwarded-For header", async () => {
  for (let i = 0; i < 100; i++) {
    const response = await fetch(PRISME_PAGEVIEWS_URL, {
      method: "POST",
      headers: {
        Origin: "http://mywebsite.localhost",
        "X-Forwarded-For": faker.internet.ip(), // ignored.
        "X-Prisme-Referrer": "http://mywebsite.localhost",
      },
    });
    if (i < 60) {
      expect(response.status).toBe(200);
    } else {
      expect(response.status).toBe(429);
    }
  }
});
