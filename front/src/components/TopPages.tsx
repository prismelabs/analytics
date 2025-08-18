import { useState } from "preact/hooks";

import Tabs from "@/components/Tabs.tsx";
import Card from "@/components/Card.tsx";
import TopBarGauge from "@/components/TopBarGauge.tsx";
import * as location from "@/signals/location.ts";
import { DataFrame, emptyDataFrame } from "@/lib/types.ts";
import useFetchJson from "@/hooks/useFetchJson.ts";

export default function () {
  const tabs = [
    { name: "Pages", resource: "top-pages", searchParam: "path" },
    {
      name: "Entry Pages",
      resource: "top-entry-pages",
      searchParam: "entry-path",
    },
    {
      name: "Exit Pages",
      resource: "top-exit-pages",
      searchParam: "exit-path",
    },
  ];

  return (
    <Card
      class="min-h-64"
      size="big"
    >
      <Tabs
        tabs={tabs.map(({ name, resource, searchParam }) => ({
          name,
          children: (
            <TopBarGauge
              key={name}
              resource={resource}
              searchParam={searchParam}
            />
          ),
        }))}
      />
    </Card>
  );
}
