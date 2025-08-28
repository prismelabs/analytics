import Tabs from "@/components/Tabs.tsx";
import Card from "@/components/Card.tsx";
import TopBarGauge from "@/components/TopBarGauge.tsx";
import { topEntryPages, topExitPages, topPages } from "@/signals/stats.ts";

export default function () {
  const tabs = [
    { name: "Pages", data: topPages, searchParam: "path" },
    { name: "Entry Pages", data: topEntryPages, searchParam: "entry-path" },
    { name: "Exit Pages", data: topExitPages, searchParam: "exit-path" },
  ];

  return (
    <Card
      class="min-h-64"
      size="big"
    >
      <Tabs
        tabs={tabs.map(({ name, data, searchParam }) => ({
          name,
          children: (
            <TopBarGauge
              key={name}
              data={data.value}
              searchParam={searchParam}
            />
          ),
        }))}
      />
    </Card>
  );
}
