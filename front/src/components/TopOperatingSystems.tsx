import Card from "@/components/Card.tsx";
import TopBarGauge from "./TopBarGauge.tsx";
import { topOperatingSystems } from "@/signals/stats.ts";

export default function () {
  return (
    <Card title="OS" size="big">
      <TopBarGauge
        data={topOperatingSystems.value}
        searchParam="os"
        transformKey={(key) => {
          if (key.trim() === "") return "Unknown";
          return key;
        }}
      />
    </Card>
  );
}
