import { useState } from "preact/hooks";
import Uplot from "@/components/Uplot.tsx";
import uPlot from "uplot";
import useStats from "@/hooks/useStats.ts";
import * as format from "@/lib/format.ts";
import { preferred as colorScheme } from "@/signals/color-scheme.ts";
import { theme } from "@/signals/trend.ts";
import { breakpoint } from "@/signals/breakpoint.ts";

export default function TimeSerie() {
  const [width, setWidth] = useState(0);
  const stroke = colorScheme.value === "light" ? "black" : "white";
  const stats = useStats();

  let filter: uPlot.Axis.Filter = (_, splits) =>
    splits.map((v, i) => i % 2 === 0 ? v : null);
  switch (breakpoint.value) {
    case "sm":
      filter = (_, splits) => splits.map((v, i) => i % 6 === 0 ? v : null);
      break;
    case "md":
      filter = (_, splits) => splits.map((v, i) => i % 3 === 0 ? v : null);
      break;
    case "lg":
      // Default.
      break;
  }

  return (
    <section class="bg-trend-background p-4 rounded">
      <div
        ref={(div) => {
          if (div) {
            const obs = new ResizeObserver((entries) => {
              for (const ent of entries) {
                setWidth(ent.contentBoxSize[0].inlineSize as number);
                break;
              }
            });
            obs.observe(div);
          }
        }}
      >
        <h2 class="font-semibold tracking-wide whitespace-nowrap text-system-muted">
          Time Series
        </h2>
        <Uplot
          options={{
            width,
            height: 312,
            legend: { show: false },
            axes: [{
              stroke,
              grid: { width: 0.05, stroke },
              ticks: { width: 0.05, stroke },
              values: (_, splits) => format.dateSerie(splits),
            }, {
              stroke,
              grid: { width: 0.05, stroke },
              ticks: { width: 0.05, stroke },
              values: (_, splits) => splits.map(format.bigNumber),
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
            plugins: [tooltipPlugin()],
          }}
          data={[stats.visitors.keys, stats.visitors.values]}
        />
      </div>
    </section>
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

        if (idx) {
          date.textContent = new Date(u.data[0][idx])
            .toLocaleDateString(navigator.language, {
              weekday: "short",
              month: "short",
              day: "2-digit",
            });
          value.textContent = format.bigNumber(u.data[1][idx] ?? 0);
        }
      },
    },
  };
}
