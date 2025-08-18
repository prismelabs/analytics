export type DataFrame<K = number> = {
  keys: Array<K>;
  values: Array<number>;
};

export const emptyDataFrame = { keys: [], values: [] };
