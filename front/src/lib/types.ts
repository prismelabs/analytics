export type DataFrame<K = number> = {
  from: number;
  to: number;
  keys: Array<K>;
  values: Array<number>;
};

export const emptyDataFrame = { from: 0, to: 0, keys: [], values: [] };
