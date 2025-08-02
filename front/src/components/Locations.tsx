import Card from "@/components/Card.tsx";
import Map from "@/components/Map.tsx";
import BarGauge from "@/components/BarGauge.tsx";

export default function Locations() {
  return (
    <Card title="Locations" size="big">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 pt-4">
        <Map class="fill-violet-200 dark:fill-indigo-300/60 stroke-system-muted md:col-span-2" />
        <BarGauge
          data={[
            { label: "France", value: 100 },
            { label: "Allemagne", value: 50 },
            { label: "Brésil", value: 0 },
          ]}
        />
      </div>
    </Card>
  );
}
