import { useState } from "preact/hooks";
import "uplot/dist/uPlot.min.css";

import Uplot from "@/components/Uplot.tsx";
import Card from "@/components/Card.tsx";

import * as format from "@/lib/format.ts";
import useStats from "@/hooks/useStats.ts";
import { theme } from "@/signals/trend.ts";

export default function Summary() {
  const {
    visitors: visitorsDF,
    pageViews: pageViewsDF,
    sessions: sessionsDF,
    sessionsDuration: sessionsDurationDF,
    liveVisitors: liveVisitorsDF,
    bounces: bouncesDF,
  } = useStats();

  const [
    visitors,
    pageViews,
    sessions,
    sessionsDuration,
    liveVisitors,
    bounces,
  ] = [
    sum(visitorsDF.values),
    sum(pageViewsDF.values),
    sum(sessionsDF.values),
    avg(sessionsDurationDF.values),
    sum(liveVisitorsDF.values),
    sum(bouncesDF.values),
  ];

  const bouncesRateDF = {
    keys: bouncesDF.keys,
    values: mul(div(bouncesDF.values, sessionsDF.values), 100),
  };

  const metrics = [
    {
      name: "Visitors",
      value: format.bigNumber(visitors),
      data: visitorsDF,
    },
    {
      name: "Sessions",
      value: format.bigNumber(sessions),
      data: sessionsDF,
    },
    {
      name: "Page views",
      value: format.bigNumber(pageViews).toString(),
      data: pageViewsDF,
    },
    {
      name: "Views per session",
      value: pageViews !== 0 && sessions !== 0
        ? (pageViews / sessions).toFixed(2)
        : "0",
      data: visitorsDF,
    },
    {
      name: "Live visitors",
      value: format.bigNumber(liveVisitors),
      data: liveVisitorsDF,
    },
    {
      name: "Bounce rate",
      value: bounces === 0
        ? "0%"
        : bouncesRateDF.values[bouncesRateDF.values.length - 1].toFixed(2) +
          "%",
      data: bouncesRateDF,
    },
    {
      name: "Avg. session duration",
      value: format.duration(sessionsDuration),
      data: sessionsDurationDF,
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
    data?: { keys: Array<number>; values: Array<number> };
  },
) {
  const hasPlot = data !== undefined && data.keys.length > 1;

  return (
    <Card
      title={name}
      size="small"
      class={`basis-0 shrink-0 grow-1 grid gap-1 items-center ${
        hasPlot ? "grid-rows-3" : "grid-rows-2"
      }`}
    >
      <p class="whitespace-nowrap text-system-fg text-center text-2xl self-start relative -top-2">
        {value}
      </p>
      {hasPlot ? <Plot {...data} /> : <div />}
    </Card>
  );
}

function Plot(
  { keys, values }: { keys: Array<number>; values: Array<number> },
) {
  const [width, setWidth] = useState(0);

  return (
    <div
      class="overflow-hidden"
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
      <Uplot
        options={{
          width,
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

const sum = (arr: Array<number>) => arr.reduce((acc, i) => acc + i, 0);
const avg = (arr: Array<number>) =>
  arr.length === 0 ? 0 : sum(arr) / arr.length;
const mul = (arr: Array<number>, scalar: number) => arr.map((v) => v * scalar);
const div = (arr: Array<number>, div: Array<number>) =>
  arr.map((v, i) => v / div[i]);
