import Card from "@/components/Card.tsx";
import Tabs from "@/components/Tabs.tsx";
import TopBarGauge from "@/components/TopBarGauge.tsx";
import {
  topReferrers,
  topUtmCampaigns,
  topUtmMediums,
  topUtmSources,
} from "@/signals/stats.ts";

export default function () {
  const tabs = [
    { name: "Referrers", data: topReferrers, searchParam: "referrer" },
    { name: "Sources", data: topUtmSources, searchParam: "utm-source" },
    { name: "Mediums", data: topUtmMediums, searchParam: "utm-medium" },
    { name: "Campaigns", data: topUtmCampaigns, searchParam: "utm-campaign" },
  ];

  return (
    <Card
      class="min-h-64"
      size="big"
    >
      <Tabs
        tabs={tabs.map(
          ({ name, data, searchParam }) => ({
            name,
            children: (
              <TopBarGauge
                key={name}
                data={data.value}
                searchParam={searchParam}
                transformKey={(key: string) => {
                  if (key.trim() === "") {
                    switch (name) {
                      case "Referrers":
                        return "Direct / None";
                      default:
                        return "Unknown";
                    }
                  }
                  if (name === "Referrers" && key.trim() === "direct") {
                    return "Direct / None";
                  }
                  return key;
                }}
              />
            ),
          }),
        )}
      />
    </Card>
  );
}
