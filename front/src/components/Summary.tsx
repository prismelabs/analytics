import { useEffect, useState } from "preact/hooks";
import "uplot/dist/uPlot.min.css";

import Uplot from "@/components/Uplot.tsx";
import Card from "@/components/Card.tsx";
import { from, to } from "@/components/TimeRangeInput.tsx"

import * as bignum from "@/lib/bignum.ts"

export default function Summary() {
  const [visitorsDF, setVisitorsDF] = useState<DataFrame>(emptyDataFrame);
  const [sessionsDF, setSessionsDF] = useState<DataFrame>(emptyDataFrame);
  const [pageViewsDF, setPageViewsDF] = useState<DataFrame>(emptyDataFrame);
  const [liveVisitorsDF, setLiveVisitorsDF] = useState<DataFrame>(
    emptyDataFrame,
  );
  const [bouncesDF, setBouncesDF] = useState<DataFrame>(emptyDataFrame);

  const [visitors, pageViews, sessions, liveVisitors, bounces] = [
    sum(visitorsDF.values),
    sum(pageViewsDF.values),
    sum(sessionsDF.values),
    sum(liveVisitorsDF.values),
    sum(bouncesDF.values),
  ];

  // @ts-ignore: vitejs magic env.
  const prismeUrl = import.meta.env.VITE_PRISME_URL;

  useEffect(() => {
    const abort = new AbortController();

    fetch(
      `${prismeUrl}/api/v1/stats/batch?metrics=visitors,sessions,pageviews,live-visitors,bounces&from=${from.value}&to=${to.value}`,
      {
        signal: abort.signal,
      },
    )
      .then((r) => r.json())
      .then((data) => {
        setVisitorsDF(data.visitors);
        setSessionsDF(data.sessions);
        setPageViewsDF(data.pageviews);
        setLiveVisitorsDF(data["live-visitors"]);
        setBouncesDF(data.bounces);
      });

    return () => abort.abort();
  }, []);

  const metrics = [
    {
      name: "Unique visitors",
      value: bignum.format( visitors),
      data: visitorsDF,
    },
    {
      name: "Unique sessions",
      value: bignum.format( sessions),
      data: visitorsDF,
    },
    {
      name: "Total page views",
      value: bignum.format(sum(pageViewsDF.values)).toString(),
      data: visitorsDF,
    },
    {
      name: "Views per session",
      value: pageViews !== 0 && sessions !== 0
        ? (pageViews / sessions).toFixed(2)
        : "0",
      data: visitorsDF,
    },
    { name: "Live visitors", value: bignum.format( liveVisitors), data: visitorsDF },
    {
      name: "Bounce rate",
      value: bounces === 0 ? "0%" : (bounces / sessions * 100).toFixed(1) + "%",
    },
    { name: "Avg. session duration", value: "59.6mins" },
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
    data?: { timestamps: Array<number>; values: Array<number> };
  },
) {
  const hasPlot = data !== undefined && data.timestamps.length > 1;

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
  { timestamps, values }: { timestamps: Array<number>; values: Array<number> },
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
              stroke: "rgb(161, 99, 208)",
              width: 1,
              points: { show: false },
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
        data={[timestamps, values]}
      />
    </div>
  );
}

const sum = (arr: Array<number>) => arr.reduce((acc, i) => acc + i, 0);

export type DataFrame = {
  timestamps: Array<number>;
  values: Array<number>;
};

const emptyDataFrame = { timestamps: [], values: [] };
