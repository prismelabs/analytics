import { expect } from "@std/expect";
import { faker } from "@faker-js/faker";

import { createClient } from "@clickhouse/client-web";
import {
  COUNTRY_CODE_REGEX,
  PRISME_OUTBOUND_LINK_EVENTS_URL,
  PRISME_PAGEVIEWS_URL,
  PRISME_VISITOR_ID_REGEX,
  TIMESTAMP_REGEX,
  UUID_V7_REGEX,
} from "../const.ts";
import { randomIpWithSession, sleep } from "../utils.ts";

const seed = new Date().getTime();
console.log("faker seed", seed);
faker.seed(seed);

Deno.test("GET request instead of POST request", async () => {
  const response = await fetch(PRISME_OUTBOUND_LINK_EVENTS_URL, {
    method: "GET",
    headers: {
      Origin: "http://mywebsite.localhost",
      "X-Forwarded-For": await randomIpWithSession("mywebsite.localhost"),
      "X-Prisme-Referrer": "http://mywebsite.localhost/",
    },
    // body: JSON.stringify({}) // GET request can't have body.
  });
  await response.body?.cancel();
  expect(response.status).toBe(405);
});

Deno.test("invalid URL in X-Prisme-Referrer header", async () => {
  const response = await fetch(PRISME_OUTBOUND_LINK_EVENTS_URL, {
    method: "POST",
    headers: {
      Origin: "http://mywebsite.localhost",
      "X-Forwarded-For": await randomIpWithSession("mywebsite.localhost"),
      "X-Prisme-Referrer": "not an url",
    },
    body: "https://www.example.com",
  });
  await response.body?.cancel();
  expect(response.status).toBe(400);
});

Deno.test("non registered domain in Origin header is rejected", async () => {
  const response = await fetch(PRISME_OUTBOUND_LINK_EVENTS_URL, {
    method: "POST",
    headers: {
      Origin: "https://example.com",
      "X-Forwarded-For": await randomIpWithSession("mywebsite.localhost"),
      "X-Prisme-Referrer": "https://example.com/",
    },
    body: "https://www.example.com",
  });
  await response.body?.cancel();
  expect(response.status).toBe(400);
});

Deno.test("relative outbound link/uri in body", async () => {
  const response = await fetch(PRISME_OUTBOUND_LINK_EVENTS_URL, {
    method: "POST",
    headers: {
      Origin: "https://mywebsite.localhost",
      "X-Forwarded-For": await randomIpWithSession("mywebsite.localhost"),
      "X-Prisme-Referrer": "https://mywebsite.localhost/",
    },
    body: "/foo/bar/baz",
  });
  await response.body?.cancel();
  expect(response.status).toBe(400);
});

Deno.test("invalid sessionless outbound link click event", async () => {
  const response = await fetch(PRISME_OUTBOUND_LINK_EVENTS_URL, {
    method: "POST",
    headers: {
      Origin: "http://mywebsite.localhost",
      // No session associated with this ip.
      "X-Forwarded-For": faker.internet.ip(),
      "X-Prisme-Referrer": "http://mywebsite.localhost/",
    },
    body: "https://www.example.com",
  });
  await response.body?.cancel();
  expect(response.status).toBe(400);
});

Deno.test("invalid Ping-From header", async () => {
  const response = await fetch(PRISME_OUTBOUND_LINK_EVENTS_URL, {
    method: "POST",
    headers: {
      Origin: "http://mywebsite.localhost",
      "X-Forwarded-For": await randomIpWithSession("mywebsite.localhost"),
      "Ping-From": "/relative/url",
      "Ping-To": "https://www.example.com",
    },
    body: "PING",
  });
  await response.body?.cancel();
  expect(response.status).toBe(400);
});

Deno.test("invalid Ping-To header", async () => {
  const response = await fetch(PRISME_OUTBOUND_LINK_EVENTS_URL, {
    method: "POST",
    headers: {
      Origin: "http://mywebsite.localhost",
      "X-Forwarded-For": await randomIpWithSession("mywebsite.localhost"),
      "Ping-From": "http://mywebsite.localhost",
      "Ping-To": "",
    },
    body: "PING",
  });
  await response.body?.cancel();
  expect(response.status).toBe(400);
});

Deno.test("invalid ping request, no session associated with Ping-From page", async () => {
  const response = await fetch(PRISME_OUTBOUND_LINK_EVENTS_URL, {
    method: "POST",
    headers: {
      Origin: "http://mywebsite.localhost",
      "X-Forwarded-For": await randomIpWithSession("mywebsite.localhost", {
        path: "/bar",
      }),
      "Ping-From": "http://mywebsite.localhost/foo",
      "Ping-To": "https://www.example.com/page1",
    },
    body: "PING",
  });
  await response.body?.cancel();
  expect(response.status).toBe(400);
});

Deno.test("valid test cases break", async () => {
  // Sleep so pageviews and identify timestamps are different for valid test
  // cases.
  // Without this sleep, getLatestXXX function may return rows from invalid test
  // cases.
  // This is not needed later as each getLatestXXX contains a 1s sleep.
  await sleep(1000);
});

Deno.test("valid outbound link click event", async () => {
  const response = await fetch(PRISME_OUTBOUND_LINK_EVENTS_URL, {
    method: "POST",
    headers: {
      Origin: "http://mywebsite.localhost",
      "X-Forwarded-For": await randomIpWithSession("mywebsite.localhost"),
      "X-Prisme-Referrer": "http://mywebsite.localhost/",
    },
    body: "https://anotherwebsite.localhost/",
  });
  await response.body?.cancel();
  expect(response.status).toBe(200);

  const data = await getLatestOutboundLinkClickEvent();

  expect(data).toMatchObject({
    session: {
      domain: "mywebsite.localhost",
      entry_path: "/",
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: "/",
      operating_system: "Other",
      browser_family: "Other",
      device: "Other",
      referrer_domain: "direct",
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: "",
      utm_medium: "",
      utm_campaign: "",
      utm_term: "",
      utm_content: "",
      version: 1,
    },
    event: {
      domain: "mywebsite.localhost",
      path: "/",
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      link: "https://anotherwebsite.localhost/",
    },
  });
});

Deno.test("valid ping request", async () => {
  const response = await fetch(PRISME_OUTBOUND_LINK_EVENTS_URL, {
    method: "POST",
    headers: {
      Origin: "http://mywebsite.localhost",
      "X-Forwarded-For": await randomIpWithSession("mywebsite.localhost"),
      "Ping-From": "http://mywebsite.localhost/foo",
      "Ping-To": "https://www.example.com/page1",
    },
    body: "PING",
  });
  await response.body?.cancel();
  expect(response.status).toBe(200);

  const data = await getLatestOutboundLinkClickEvent();

  expect(data).toMatchObject({
    session: {
      domain: "mywebsite.localhost",
      entry_path: "/",
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: "/",
      operating_system: "Other",
      browser_family: "Other",
      device: "Other",
      referrer_domain: "direct",
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: "",
      utm_medium: "",
      utm_campaign: "",
      utm_term: "",
      utm_content: "",
      version: 1,
    },
    event: {
      domain: "mywebsite.localhost",
      path: "/", // path of session and not path of Ping-From
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      link: "https://www.example.com/page1",
    },
  });
});

Deno.test("concurrent pageview and outbound link click event", async () => {
  const ipAddr = faker.internet.ip();

  await Promise.all([
    // Click events first.
    fetch(PRISME_OUTBOUND_LINK_EVENTS_URL, {
      method: "POST",
      headers: {
        Origin: "http://mywebsite.localhost",
        "X-Forwarded-For": ipAddr,
        "X-Prisme-Referrer": "http://mywebsite.localhost/",
      },
      body: "https://anotherwebsite.localhost/",
    }),
    // Pageview concurrently.
    // This pageview will create session that will be used for both events.
    fetch(PRISME_PAGEVIEWS_URL, {
      method: "POST",
      headers: {
        Origin: "https://mywebsite.localhost",
        "X-Forwarded-For": ipAddr,
        "X-Prisme-Referrer": "https://mywebsite.localhost/",
      },
    }),
  ]).then((results) =>
    results.forEach(async (resp) => {
      await resp.body?.cancel();
      expect(resp.status).toBe(200);
    })
  );

  const data = await getLatestOutboundLinkClickEvent();

  expect(data).toMatchObject({
    session: {
      domain: "mywebsite.localhost",
      entry_path: "/",
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: "/",
      operating_system: "Other",
      browser_family: "Other",
      device: "Other",
      referrer_domain: "direct",
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: "",
      utm_medium: "",
      utm_campaign: "",
      utm_term: "",
      utm_content: "",
      version: 1,
    },
    event: {
      domain: "mywebsite.localhost",
      path: "/",
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      link: "https://anotherwebsite.localhost/",
    },
  });
});

Deno.test("valid click event without X-Prisme-Referrer", async () => {
  const response = await fetch(PRISME_OUTBOUND_LINK_EVENTS_URL, {
    method: "POST",
    headers: {
      Origin: "http://mywebsite.localhost",
      "X-Forwarded-For": await randomIpWithSession("mywebsite.localhost"),
      Referer: "http://mywebsite.localhost/",
    },
    body: "https://anotherwebsite.localhost/",
  });
  await response.body?.cancel();
  expect(response.status).toBe(200);

  const data = await getLatestOutboundLinkClickEvent();

  expect(data).toMatchObject({
    session: {
      domain: "mywebsite.localhost",
      entry_path: "/",
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: "/",
      operating_system: "Other",
      browser_family: "Other",
      device: "Other",
      referrer_domain: "direct",
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: "",
      utm_medium: "",
      utm_campaign: "",
      utm_term: "",
      utm_content: "",
      version: 1,
    },
    event: {
      domain: "mywebsite.localhost",
      path: "/",
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      link: "https://anotherwebsite.localhost/",
    },
  });
});

Deno.test("valid click event without trailing slash in referrer", async () => {
  const response = await fetch(PRISME_OUTBOUND_LINK_EVENTS_URL, {
    method: "POST",
    headers: {
      Origin: "http://mywebsite.localhost",
      "X-Forwarded-For": await randomIpWithSession("mywebsite.localhost"),
      Referer: "http://mywebsite.localhost",
    },
    body: "https://anotherwebsite.localhost/",
  });
  await response.body?.cancel();
  expect(response.status).toBe(200);

  const data = await getLatestOutboundLinkClickEvent();

  expect(data).toMatchObject({
    session: {
      domain: "mywebsite.localhost",
      entry_path: "/",
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: "/",
      operating_system: "Other",
      browser_family: "Other",
      device: "Other",
      referrer_domain: "direct",
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: "",
      utm_medium: "",
      utm_campaign: "",
      utm_term: "",
      utm_content: "",
      version: 1,
    },
    event: {
      domain: "mywebsite.localhost",
      path: "/",
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      link: "https://anotherwebsite.localhost/",
    },
  });
});

Deno.test("valid click event with Windows + Chrome user agent", async () => {
  const userAgent =
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.3";

  const response = await fetch(PRISME_OUTBOUND_LINK_EVENTS_URL, {
    method: "POST",
    headers: {
      Origin: "http://mywebsite.localhost",
      "X-Forwarded-For": await randomIpWithSession("mywebsite.localhost", {
        userAgent,
      }),
      Referer: "http://mywebsite.localhost",
      "User-Agent": userAgent,
    },
    body: "https://anotherwebsite.localhost/",
  });
  await response.body?.cancel();
  expect(response.status).toBe(200);

  const data = await getLatestOutboundLinkClickEvent();

  expect(data).toMatchObject({
    session: {
      domain: "mywebsite.localhost",
      entry_path: "/",
      exit_timestamp: expect.stringMatching(TIMESTAMP_REGEX),
      exit_path: "/",
      operating_system: "Windows",
      browser_family: "Chrome",
      device: "Other",
      referrer_domain: "direct",
      country_code: expect.stringMatching(COUNTRY_CODE_REGEX),
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      utm_source: "",
      utm_medium: "",
      utm_campaign: "",
      utm_term: "",
      utm_content: "",
      version: 1,
    },
    event: {
      domain: "mywebsite.localhost",
      path: "/",
      visitor_id: expect.stringMatching(PRISME_VISITOR_ID_REGEX),
      session_uuid: expect.stringMatching(UUID_V7_REGEX),
      link: "https://anotherwebsite.localhost/",
    },
  });
});

// deno-lint-ignore no-explicit-any
async function getLatestOutboundLinkClickEvent(): Promise<any> {
  // Wait for clickhouse to ingest batch.
  await sleep(1000);

  const client = createClient({
    url: "http://clickhouse.localhost:8123",
    username: "clickhouse",
    password: "password",
    database: "prisme",
  });

  const sessions = await client.query({
    query: "SELECT * FROM sessions ORDER BY exit_timestamp DESC LIMIT 1",
  });
  // deno-lint-ignore no-explicit-any
  const session = await sessions.json().then((r: any) => r.data[0]);
  expect(session.sign).toBe(1);
  delete session.sign;

  const clickEvents = await client.query({
    query: `SELECT * FROM outbound_link_clicks WHERE visitor_id = '${session
      .visitor_id as string}' ORDER BY timestamp DESC LIMIT 1`,
  });
  // deno-lint-ignore no-explicit-any
  const clickEvent = await clickEvents.json().then((r: any) => r.data[0]);
  if (clickEvent === null || clickEvent === undefined) return null;
  expect(clickEvent.visitor_id).toBe(session.visitor_id);
  expect(clickEvent.session_uuid).toBe(session.session_uuid);

  return { event: clickEvent, session };
}
