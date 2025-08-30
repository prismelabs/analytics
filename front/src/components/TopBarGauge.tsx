import { DataFrame } from "@/lib/types.ts";
import * as location from "@/signals/location.ts";
import BarGauge from "./BarGauge.tsx";
import { useEffect } from "preact/hooks";
import { useSignal } from "@preact/signals";

export default function (
  { data, searchParam, transformKey, onMouseEnter, onMouseLeave }: {
    data: DataFrame<string>;
    searchParam: string;
    transformKey?: (_: string) => string;
    onMouseEnter?: (_: { label: string; value: number }) => void;
    onMouseLeave?: (_: { label: string; value: number }) => void;
  },
) {
  const multiple = useSignal(false);
  const selected = useSignal<Record<string, boolean>>({});
  const updateFilters = () => {
    if (multiple.value) return;

    location.update((url) => {
      const filter = Object.fromEntries(
        (url.searchParams.get(searchParam) ?? "")
          .split(",")
          .filter((item) => !!item)
          .map((i) => [i, true]),
      );
      for (const s in selected.value) {
        filter[s] = true;
      }
      selected.value = {};

      url.searchParams.set(searchParam, Object.keys(filter).join(","));
    });
  };

  useEffect(() => {
    const listener = (ev: KeyboardEvent) => {
      if (ev.key !== "Control") return;

      switch (ev.type) {
        case "keyup":
          multiple.value = false;
          updateFilters();
          break;
        case "keydown":
          multiple.value = true;
          break;
      }
    };

    document.addEventListener("keydown", listener);
    document.addEventListener("keyup", listener);

    return () => {
      document.removeEventListener("keydown", listener);
      document.removeEventListener("keyup", listener);
    };
  }, []);

  return (
    <BarGauge
      selected={selected}
      data={data.keys.map((k, i) => {
        const label = transformKey ? transformKey(k) : k;
        return {
          label,
          value: data.values[i],
          onClick: searchParam
            ? () => {
              selected.value[label] = true;
              updateFilters();
            }
            : undefined,
          onMouseEnter: onMouseEnter,
          onMouseLeave: onMouseLeave,
        };
      })}
    />
  );
}
