import * as location from "@/signals/location.ts";
import BarGauge from "@/components/BarGauge.tsx";
import { useEffect } from "preact/hooks";
import { computed, useSignal } from "@preact/signals";
import { FetchedStat } from "@/signals/stats.ts";
import { isError } from "@/lib/types.ts";

export default function (
  { data, searchParam, transformKey, onMouseEnter, onMouseLeave }: {
    data: FetchedStat<string>;
    searchParam: string;
    transformKey?: (_: string) => string;
    onMouseEnter?: (_: { label: string; value: number }) => void;
    onMouseLeave?: (_: { label: string; value: number }) => void;
  },
) {
  const multiple = useSignal(false);
  const selected = useSignal<Record<string, boolean>>({});
  const selectedLabels = computed(() =>
    Object.fromEntries(
      Object.entries(selected.value).map((
        [k, v],
      ) => [transformKey ? transformKey(k) : k, v]),
    )
  );

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

  if (isError(data)) {
    return (
      <span class="font-bold text-page-fg text-center flex flex-row min-h-56 items-center justify-center">
        {data.error}
      </span>
    );
  }

  return (
    <BarGauge
      selected={selectedLabels}
      data={data.ok.keys.map((k, i) => {
        const label = transformKey ? transformKey(k) : k;
        return {
          label,
          value: data.ok.values[i],
          onClick: searchParam
            ? () => {
              selected.value = { ...selected.value, [k]: true };
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
