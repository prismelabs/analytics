import Card from "@/components/Card.tsx";
import TopBarGauge from "@/components/TopBarGauge.tsx";
import { topBrowsers } from "@/signals/stats.ts";

export default function () {
  return (
    <Card title="Browser" size="big">
      <TopBarGauge
        data={topBrowsers.value}
        searchParam="browser"
        transformKey={(str: string) => {
          if (str.trim() === "") return "Unknown";
          return str;
        }}
      />
    </Card>
  );
}
