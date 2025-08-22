import { useState } from "preact/hooks";
import { DataFrame, emptyDataFrame } from "@/lib/types.ts";
import * as location from "@/signals/location.ts";
import useFetchJson from "@/hooks/useFetchJson.ts";
import BarGauge from "./BarGauge.tsx";

export default function (
  { resource, searchParam, transformKey, onMouseEnter, onMouseLeave }: {
    resource: string;
    searchParam?: string;
    transformKey?: (_: string) => string;
    onMouseEnter?: (_: { label: string; value: number }) => void;
    onMouseLeave?: (_: { label: string; value: number }) => void;
  },
) {
  // deno-lint-ignore no-explicit-any
  const [top, setTop] = useState<DataFrame<any>>(emptyDataFrame);

  const search = location.current.value.search;
  const loc = location.current.value.toString();
  const useFetchStats = (stat: string) =>
    useFetchJson<DataFrame<string>>(
      `/api/v1/stats/${stat}${search}`,
      [loc],
    );

  useFetchStats(resource).then(setTop);

  const data = [];
  for (let j = 0; j < top.keys.length; j++) {
    const k = top.keys[j];
    const v = top.values[j];
    data.push({
      label: transformKey ? transformKey(k) : k,
      value: v,
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
    });
  }

  return (
    <BarGauge
      data={data}
    />
  );
}
