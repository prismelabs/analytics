export type TimeRange = {
  from: string;
  to: string;
};

const strings: Record<string, string> = {
  "now-7d:now": "Last 7 days",
};

export function toString(tr: TimeRange): string {
  const key = tr.from + ":" + tr.to;
  return strings[key] ?? "TODO";
}
