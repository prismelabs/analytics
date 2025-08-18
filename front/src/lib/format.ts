export function bigNumber(num: number): string {
  const absNum = Math.abs(num);
  if (absNum < 1100) return num.toString();

  const suffixes = ["K", "M", "B", "T"];
  const suffixIndex = Math.floor(Math.log10(absNum) / 3) - 1;
  const shortNum = (num / Math.pow(1000, suffixIndex + 1)).toFixed(1);

  return `${shortNum}${suffixes[suffixIndex]}`;
}

export function duration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  const formattedParts = [];
  if (hours > 0) {
    formattedParts.push(`${hours.toFixed(0)}h`);
  }
  if (minutes > 0) {
    formattedParts.push(`${minutes.toFixed(0)}m`);
  }
  if (secs > 0 || formattedParts.length === 0) {
    formattedParts.push(`${secs.toFixed(0)}s`);
  }

  return formattedParts.join(" ");
}

export function dateSerie(
  ts: Array<number | null>,
  locale?: string,
): Array<string> {
  if (ts.length === 0) return [];

  const dates = ts.map((ts) => ts !== null ? new Date(ts) : null);
  locale = locale ? locale : navigator.language;

  const nonNullDates = dates.filter((d) => d !== null);
  const firstDate = nonNullDates[0];
  const lastDate = nonNullDates[nonNullDates.length - 1];

  if (
    firstDate?.getFullYear() === lastDate?.getFullYear() &&
    firstDate?.getMonth() === lastDate?.getMonth() &&
    firstDate?.getDate() === lastDate?.getDate()
  ) {
    return dates.map((d) =>
      d === null ? "" : Intl.DateTimeFormat(locale, {
        dateStyle: "short",
        timeStyle: "short",
      }).format(d)
    );
  } else {
    return dates.map((d) =>
      d === null ? "" : Intl.DateTimeFormat(locale, {
        dateStyle: "short",
        timeStyle: "short",
      }).format(d)
    );
  }
}
