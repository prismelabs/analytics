export type TimeRange = {
  from: string;
  to: string;
};

export type DataFrame<K = number> = {
  from: number;
  to: number;
  keys: Array<K>;
  values: Array<number>;
};

export const emptyDataFrame = { from: 0, to: 0, keys: [], values: [] };

export type Result<Ok, Err> = {
  ok: Ok;
  error: null;
} | {
  ok: null;
  error: Err;
};

export function ok<Ok, Err = string>(ok: Ok): Result<Ok, Err> {
  return { ok, error: null };
}

export function error<Ok, Err = string>(error: Err): Result<Ok, Err> {
  return { ok: null, error };
}

export function isOk<Ok, Err>(
  r: Result<Ok, Err>,
): r is { ok: Ok; error: null } {
  return (r.ok !== null);
}

export function isError<Ok, Err>(
  r: Result<Ok, Err>,
): r is { ok: null; error: Err } {
  return (r.error !== null);
}
