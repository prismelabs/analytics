import { useRef } from "preact/hooks";
import Uplot from "@/components/Uplot.tsx";
import Card from "@/components/Card.tsx";
import * as format from "@/lib/format.ts";
import { theme } from "@/signals/trend.ts";
import useSizeOf from "@/hooks/useSizeOf.ts";
import { Signal, signal } from "@preact/signals";
import { DataFrame } from "@/lib/types.ts";
import {
  avgSessionsDuration,
  liveVisitors,
  pageViews,
  sessions,
  sessionsDuration,
  Stat,
  totalLiveVisitors,
  totalPageViews,
  totalSessions,
  totalVisitors,
  viewsPerSessions,
  visitors,
} from "@/signals/stats.ts";
import LoadingBar from "./LoadingBar.tsx";

export const selectedTimeSerie = signal<Signal<DataFrame | null>>(visitors);

export default function Summary() {
  const metrics = [
    {
      name: "Visitors",
      value: format.bigNumber(totalVisitors.value),
      data: visitors,
    },
    {
      name: "Sessions",
      value: format.bigNumber(totalSessions.value),
      data: sessions,
    },
    {
      name: "Page views",
      value: format.bigNumber(totalPageViews.value),
      data: pageViews,
    },
    {
      name: "Views per session",
      value: `${viewsPerSessions.value.toFixed(2)}%`,
    },
    {
      name: "Live visitors",
      value: format.bigNumber(totalLiveVisitors.value),
      data: liveVisitors,
    },
    {
      name: "Bounce rate",
      value: "0%",
    },
    {
      name: "Avg. sessions duration",
      value: format.duration(avgSessionsDuration.value),
      data: sessionsDuration,
    },
  ];

  return (
    <>
      <section class="flex flex-wrap justify-center gap-4">
        {metrics.map((m, i) => <Metric key={i} {...m} />)}
      </section>
    </>
  );
}

function Metric(
  { name, value, data }: {
    name: string;
    value: string;
    data?: Signal<Stat>;
  },
) {
  const hasPlot = data !== undefined && data.value.keys.length > 1;
  const selected = selectedTimeSerie.value === data;

  return (
    <Card
      title={name}
      size="small"
      class={`basis-0 shrink-0 grow-1 min-w-max grid gap-1 items-center border-x-4 border-x-transparent px-3 relative ${
        hasPlot ? "grid-rows-3 cursor-pointer" : "grid-rows-2"
      } ${selected ? "border-l-trend-primary" : ""}`}
      onClick={hasPlot ? () => selectedTimeSerie.value = data : undefined}
    >
      {data !== undefined
        ? <LoadingBar loading={data.value.loading} class="absolute top-0" />
        // deno-lint-ignore jsx-no-useless-fragment
        : <></>}
      <p class="whitespace-nowrap text-system-fg text-center text-2xl self-start relative -top-2">
        {value}
      </p>
      {hasPlot ? <Plot {...data.value} /> : <div />}
    </Card>
  );
}

function Plot(
  { keys, values }: { keys: Array<number>; values: Array<number> },
) {
  const plotRef = useRef(null);
  const size = useSizeOf(plotRef);

  return (
    <div
      class="overflow-hidden"
      ref={plotRef}
    >
      <Uplot
        options={{
          width: size.width,
          height: 36,
          cursor: { show: false },
          legend: { show: false },
          axes: [{ show: false }, { show: false }],
          series: [
            {},
            {
              show: true,
              stroke: theme.value.primary,
              width: 1,
              points: { show: false },
              fill: theme.value["primary-light"] + "88",
            },
          ],
        }}
        data={[keys, values]}
      />
    </div>
  );
}
