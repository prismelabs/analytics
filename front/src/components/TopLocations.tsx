import { useState } from "preact/hooks";

import Card from "@/components/Card.tsx";
import Map from "@/components/Map.tsx";
import TopBarGauge from "@/components/TopBarGauge.tsx";
import { countryClass, countryCode, countryName } from "@/lib/countries.ts";
import { topCountries } from "@/signals/stats.ts";

export default function Locations() {
  const [country, setCountry] = useState<string | null>(null);

  return (
    <Card title="Locations" size="big">
      <div
        class="grid grid-cols-1 lg:grid-cols-3 gap-4 pt-4 items-center"
        ref={(d) => {
          if (!d) return;

          const activeClass = "fill-trend-primary";
          d.querySelectorAll(
            [".country", activeClass].join("."),
          ).forEach((el) => el.classList.remove(activeClass));

          if (
            country && countryCode[country] &&
            countryClass[countryCode[country]]
          ) {
            const selector = [
              ".country",
              ...countryClass[countryCode[country]].split(" "),
            ].join(".");
            d.querySelectorAll(selector).forEach((el) => {
              el.classList.add(activeClass);
            });
          }
        }}
      >
        <Map class="fill-trend-primary-light stroke-system-muted md:col-span-2" />
        <TopBarGauge
          data={topCountries.value}
          searchParam="country"
          transformKey={(k) => countryName[k] ?? "Unknown"}
          onMouseEnter={({ label }) => setCountry(label)}
          onMouseLeave={() => setCountry(null)}
        />
      </div>
    </Card>
  );
}
