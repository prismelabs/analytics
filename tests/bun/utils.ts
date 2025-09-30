import { expect } from "@std/expect";
import { faker } from "@faker-js/faker";
import { PRISME_PAGEVIEWS_URL } from "./const.ts";

export async function randomIpWithSession(
  domain: string,
  options?: Partial<{ userAgent: string; visitorId: string; path: string }>,
): Promise<string> {
  const ip = faker.internet.ip();
  const headers: HeadersInit = {
    Origin: `http://${domain}`,
    "X-Forwarded-For": ip,
    "X-Prisme-Referrer": `http://${domain}/`,
  };
  if (options?.userAgent !== undefined) {
    headers["User-Agent"] = options.userAgent;
  }
  if (options?.visitorId !== undefined) {
    headers["X-Prisme-Visitor-Id"] = options.visitorId;
  }
  if (options?.path !== undefined) {
    headers["X-Prisme-Referrer"] = `http://${domain}${options.path}`;
  }

  const response = await fetch(PRISME_PAGEVIEWS_URL, {
    method: "POST",
    headers,
  });
  await response.body?.cancel();

  expect(response.status).toBe(200);

  return ip;
}

export function sleep(ms: number): Promise<void> {
  return new Promise((res) => setTimeout(res, ms));
}
