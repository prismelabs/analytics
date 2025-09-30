import { expect } from "@std/expect";

Deno.test("X-Forwarded-For is used when present", async () => {
  const sourceIp = "10.127.42.1";

  // Generate an access log.
  await fetch("http://prisme.localhost/", {
    headers: {
      "X-Forwarded-For": sourceIp,
    },
  });

  // Read access log file.
  const text = await Deno.readTextFile("/prisme_logs/access.log");
  const lines = text.split("\n").filter((l) => l !== "");
  const lastLogLine = lines[lines.length - 1];
  const lastLog = JSON.parse(lastLogLine);

  expect(lastLog.source_ip).toBe(sourceIp);
});

Deno.test("Real IP is used when X-Forwarded-For is missing", async () => {
  // Generate an access log.
  await fetch("http://prisme.localhost/");

  // Read access log file.
  const text = await Deno.readTextFile("/prisme_logs/access.log");
  const lines = text.split("\n").filter((l) => l !== "");
  const lastLogLine = lines[lines.length - 1];
  const lastLog = JSON.parse(lastLogLine);

  expect(lastLog.source_ip).toBe("");
});
