const alpha = /[a-zA-Z]/;
const digit = /[0-9]/;
const duration = /^\d+(y|M|d|h|m|s)$/;

export function normalize(expr: string): string {
  return expr.split(" ").join("");
}

export function parse(expr: string): Date {
  let sign = 1;
  let skip = 0;
  let date: null | Date = null;

  for (let i = 0; i < expr.length; i++) {
    const c = expr[i];

    if (skip > 0) {
      skip--;
      continue;
    }

    if (c.match(alpha)) {
      if (date !== null) throw new InvalidSyntaxError();

      const { date: refDate, ref } = parseRef(expr);
      date = refDate;
      skip = ref.length - 1;
      continue;
    }

    if (c == "-" || c == "+") {
      if (date === null) throw new InvalidSyntaxError();

      if (c == "-") sign = -1;
      continue;
    }

    if (c.match(digit)) {
      if (date === null) {
        return new Date(expr);
      } else if (expr.slice(i).match(duration)) {
        const n = Number.parseInt(expr.slice(i, expr.length - 1), 10) * sign;

        switch (expr[expr.length - 1]) {
          case "y":
            date.setFullYear(date.getFullYear() + n);
            break;
          case "M":
            date.setMonth(date.getMonth() + n);
            break;
          case "d":
            date.setDate(date.getDate() + n);
            break;
          case "h":
            date.setHours(date.getHours() + n);
            break;
          case "m":
            date.setMinutes(date.getMinutes() + n);
            break;
          case "s":
            date.setSeconds(date.getSeconds() + n);
            break;
          default:
            throw new InvalidSyntaxError();
        }

        break;
      }
    }

    throw new InvalidSyntaxError();
  }

  if (date === null) throw new InvalidSyntaxError();
  return date;
}

const refs: Record<string, () => Date> = {
  "now": () => new Date(),
};

export function parseRef(expr: string): { date: Date; ref: string } {
  let ref = "";
  for (const c of expr) {
    if (!c.match(alpha)) break;
    ref += c;
  }

  const dateFn = refs[ref];
  if (!dateFn) throw new InvalidSyntaxError();

  return { date: dateFn(), ref };
}

class InvalidSyntaxError extends Error {
  constructor() {
    super("invalid time expression syntax");
  }
}
