import { useRef } from "preact/hooks";
import Uplot from "@/components/Uplot.tsx";
import uPlot from "uplot";
import * as format from "@/lib/format.ts";
import { preferred as colorScheme } from "@/signals/color-scheme.ts";
import { theme } from "@/signals/trend.ts";
import useSizeOf from "@/hooks/useSizeOf.ts";
import { measureText } from "@/lib/canvas.ts";
import { selectedTimeRange } from "@/components/TimeRangeInput.tsx";
import Card from "@/components/Card.tsx";
import { selectedTimeSerie } from "@/components/Summary.tsx";
import { sessionsDuration } from "@/signals/stats.ts";

const max = (acc: number, v: number) => v > acc ? v : acc;

export default function TimeSerie() {
  const plotRef = useRef(null);
  const size = useSizeOf(plotRef);

  const stroke = colorScheme.value === "light" ? "black" : "white";
  const stat = selectedTimeSerie.value.value;
  if (!stat) return;

  const formatY = selectedTimeSerie.value === sessionsDuration
    ? format.duration
    : format.bigNumber;

  return (
    <Card size="big">
      <div ref={plotRef}>
        <h2 class="font-semibold tracking-wide whitespace-nowrap text-system-muted">
          Time Series
        </h2>
        {stat.keys.length === 0
          ? (
            <div class="flex h-[312px] items-center text-center">
              <p class="w-full font-bold">No data</p>
            </div>
          )
          : (
            <Uplot
              key={stat}
              options={{
                width: size.width,
                height: 312,
                legend: { show: false },
                scales: {
                  // x: { range: [stat.from * 1000, stat.to * 1000] },
                  y: { range: [0, stat.values.reduce(max, 0)] },
                },
                axes: [{
                  stroke,
                  grid: { width: 0.05, stroke },
                  ticks: { width: 0.05, stroke },
                  font: "400 12px Arial",
                  space: (
                    _self: uPlot,
                    _axisIdx: number,
                    scaleMin: number,
                    scaleMax: number,
                    _plotDim: number,
                    _formatValue?: (value: unknown) => string,
                  ) => {
                    const sample = format.dateSerie([scaleMin, scaleMax])[0];
                    return measureText(sample, 12).width + 8;
                  },
                  values: (_, splits) => format.dateSerie(splits),
                }, {
                  stroke,
                  grid: { width: 0.05, stroke },
                  ticks: { width: 0.05, stroke },
                  values: (_, splits) => splits.map(formatY),
                  size: (
                    _: uPlot,
                    values: string[],
                  ) => {
                    if (!values) return 0;
                    const w = values.reduce(
                      (acc, str) => Math.max(measureText(str, 12).width, acc),
                      0,
                    );
                    return Math.ceil(w) + 24;
                  },
                }],
                series: [
                  {},
                  {
                    show: true,
                    stroke: theme.value.primary,
                    width: 1,
                    fill: theme.value["primary-light"] + "88",
                  },
                ],
                plugins: [tooltipPlugin(), selectTimeRangePlugin],
              }}
              data={[stat.keys, stat.values]}
            />
          )}
      </div>
    </Card>
  );
}

function tooltipPlugin(): uPlot.Plugin {
  let cursortt: HTMLDivElement;
  let date: HTMLParagraphElement;
  let value: HTMLParagraphElement;

  return {
    hooks: {
      init: (u) => {
        const over = u.over;
        cursortt = document.createElement("div");
        cursortt.className = "tooltip";
        cursortt.style.display = "none";
        cursortt.style.pointerEvents = "none";
        cursortt.style.position = "absolute";
        cursortt.style.background =
          "color-mix(in oklab, var(--color-trend-background) 90%, transparent)";
        cursortt.style.padding = "calc(var(--spacing) * 2)";
        cursortt.style.translate = "0% -100%";
        cursortt.style.color = "var(--system-fg)";
        cursortt.style.borderRadius = "0.25rem";
        cursortt.style.width = "max-content";
        cursortt.style.fontSize = "0.90rem";
        // tailwind .shadow-sm
        cursortt.style.boxShadow =
          "rgba(0, 0, 0, 0) 0px 0px 0px 0px, rgba(0, 0, 0, 0) 0px 0px 0px 0px, rgba(0, 0, 0, 0) 0px 0px 0px 0px, rgba(0, 0, 0, 0) 0px 0px 0px 0px, rgba(0, 0, 0, 0.1) 0px 1px 3px 0px, rgba(0, 0, 0, 0.1) 0px 1px 2px -1px";

        date = document.createElement("p");
        value = document.createElement("p");

        value.style.color = "var(--page-fg)";

        cursortt.appendChild(date);
        cursortt.appendChild(value);

        over.appendChild(cursortt);

        over.addEventListener("mouseenter", () => {
          cursortt.style.display = "block";
        });
        over.addEventListener("mouseleave", () => {
          cursortt.style.display = "none";
        });
      },
      setCursor: (u) => {
        const { left, top, idx } = u.cursor;
        cursortt.style.left = left + "px";
        cursortt.style.top = top + "px";

        const rect = u.over.getBoundingClientRect();
        if (
          rect.left + (left ?? 0) +
              cursortt.getBoundingClientRect().width >
            globalThis.innerWidth
        ) {
          cursortt.style.translate = "-100% -100%";
        } else {
          cursortt.style.translate = "0% -100%";
        }

        if (idx !== null && idx !== undefined) {
          date.textContent = Intl.DateTimeFormat(undefined, {
            dateStyle: "short",
            timeStyle: "short",
          }).format(new Date(u.data[0][idx]));
          value.textContent = format.bigNumber(u.data[1][idx] ?? 0);
        }
      },
    },
  };
}

const selectTimeRangePlugin: uPlot.Plugin = {
  hooks: {
    setSelect: (u) => {
      const min = u.posToVal(u.select.left, "x");
      const max = u.posToVal(u.select.left + u.select.width, "x");

      const from = new Date(min);
      const to = new Date(max);

      selectedTimeRange.value = {
        from: from.toISOString(),
        to: to.toISOString(),
      };
    },
  },
};
