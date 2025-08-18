import Card from "@/components/Card.tsx";
import TopBarGauge from "@/components/TopBarGauge.tsx";

export default function () {
  return (
    <Card title="Browser" size="big">
      <TopBarGauge
        resource="top-browsers"
        searchParam="browser"
        transformKey={(str: string) => {
          if (str.trim() === "") return "Unknown";
          return str;
        }}
      />
    </Card>
  );
}
