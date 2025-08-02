import { useState } from "preact/hooks";
import Uplot from "@/components/Uplot.tsx";
import useTheme from "@/hooks/useTheme.ts";

export default function TimeSerie() {
  const theme = useTheme();
  const [width, setWidth] = useState(0);
  const stroke = theme === "light" ? "black" : "white";

  return (
    <section class="bg-system-bg p-4 rounded">
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
            }, {
              stroke,
              grid: { width: 0.05, stroke },
              ticks: { width: 0.05, stroke },
            }],
            series: [
              {},
              {
                show: true,
                stroke: "rgb(161, 99, 208)",
                width: 1,
                fill: "rgba(72, 143, 220, 0.25)",
              },
              // {
              //   show: true,
              //   stroke: "rgb(161, 99, 208)",
              //   width: 1,
              //   fill: "rgba(161, 99, 208, 0.25)",
              // },
            ],
          }}
          data={[
            [1546300800, 1546387200, 1546489199], // x-values (timestamps)
            [35, 71, 22], // y-values (series 1)
          ]}
        />
      </div>
    </section>
  );
}
