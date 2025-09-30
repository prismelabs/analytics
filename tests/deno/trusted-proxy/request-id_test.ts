import { expect } from "@std/expect";
import { UUID_V4_REGEX } from "../const.ts";

Deno.test("X-Request-Id is used when present", async () => {
  const requestId = "foo";

  // Generate an access log.
  await fetch("http://prisme.localhost/", {
    headers: {
      "X-Request-Id": requestId,
    },
  });

  // Read access log file.
  const text = await Deno.readTextFile("/prisme_logs/access.log");
  const lines = text.split("\n").filter((l) => l !== "");
  const lastLogLine = lines[lines.length - 1];
  const lastLog = JSON.parse(lastLogLine);

  expect(lastLog.request_id).toBe(requestId);
});

Deno.test("Random UUID v4 is used when X-Request-Id is missing", async () => {
  // Generate an access log.
  await fetch("http://prisme.localhost/");

  // Read access log file.
  const text = await Deno.readTextFile("/prisme_logs/access.log");
  const lines = text.split("\n").filter((l) => l !== "");
  const lastLogLine = lines[lines.length - 1];
  const lastLog = JSON.parse(lastLogLine);

  expect(lastLog.request_id).toMatch(UUID_V4_REGEX);
});
