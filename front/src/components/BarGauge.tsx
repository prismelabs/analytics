import * as format from "@/lib/format.ts";
import { Signal } from "@preact/signals";

export default function BarGauge(
  { data, selected }: {
    data: Array<
      {
        label: string;
        value: number;
        onClick?: (_: { label: string; value: number }) => void;
        onMouseEnter?: (_: { label: string; value: number }) => void;
        onMouseLeave?: (_: { label: string; value: number }) => void;
      }
    >;
    selected: Signal<Record<string, boolean>>;
  },
) {
  data.sort((a, b) => b.value - a.value);

  const min = data.length > 0 ? data[data.length - 1].value : 0;
  const max = data.length > 0 ? data[0].value : 0;

  return (data.length === 0
    ? (
      <span class="font-bold text-page-fg text-center flex flex-row min-h-56 items-center justify-center">
        No data
      </span>
    )
    : (
      <ul class="flex flex-col gap-2 text-system-fg">
        {data.map((d, i) => (
          <li
            key={i}
            class={`rounded hover:bg-trend-page cursor-pointer text-select-none ${
              selected.value[d.label] ? "bg-trend-page" : ""
            }`}
            onClick={d.onClick ? () => d.onClick!(d) : undefined}
            onMouseEnter={d.onMouseEnter ? () => d.onMouseEnter!(d) : undefined}
            onMouseLeave={d.onMouseLeave ? () => d.onMouseLeave!(d) : undefined}
          >
            <div
              class="flex justify-between py-1 px-2 overflow-hidden"
              title={d.label}
            >
              <span class="text-nowrap text-ellipsis overflow-hidden marquee">
                {d.label}
              </span>
              <span class="min-w-max">{format.bigNumber(d.value)}</span>
            </div>
            <Bar percentage={(d.value - min) / (max - min) * 100} />
          </li>
        ))}
      </ul>
    ));
}

function Bar({ percentage }: { percentage: number }) {
  return (
    <div class="h-0.5">
      <div
        style={`width: ${percentage}%`}
        class="h-0.5 bg-trend-primary"
      >
      </div>
    </div>
  );
}
