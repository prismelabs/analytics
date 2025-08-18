import Card from "@/components/Card.tsx";
import TopBarGauge from "./TopBarGauge.tsx";

export default function () {
  return (
    <Card title="OS" size="big">
      <TopBarGauge
        resource="top-operating-systems"
        searchParam="os"
        transformKey={(key) => {
          if (key.trim() === "") return "Unknown";
          return key;
        }}
      />
    </Card>
  );
}
