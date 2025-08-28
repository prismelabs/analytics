import { DataFrame } from "@/lib/types.ts";
import * as location from "@/signals/location.ts";
import BarGauge from "./BarGauge.tsx";

export default function (
  { data, searchParam, transformKey, onMouseEnter, onMouseLeave }: {
    data: DataFrame<string>;
    searchParam?: string;
    transformKey?: (_: string) => string;
    onMouseEnter?: (_: { label: string; value: number }) => void;
    onMouseLeave?: (_: { label: string; value: number }) => void;
  },
) {
  return (
    <BarGauge
      data={data.keys.map((k, i) => ({
        label: transformKey ? transformKey(k) : k,
        value: data.values[i],
        onClick: searchParam
          ? () =>
            location.update((url) => {
              const filter = (url.searchParams.get(searchParam) ?? "")
                .split(",").filter((item) => !!item);
              filter.push(k);

              url.searchParams.set(searchParam, filter.join(","));
            })
          : undefined,
        onMouseEnter: onMouseEnter,
        onMouseLeave: onMouseLeave,
      }))}
    />
  );
}
