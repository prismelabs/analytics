import Card from "@/components/Card.tsx";
import Tabs from "@/components/Tabs.tsx";
import TopBarGauge from "@/components/TopBarGauge.tsx";

export default function () {
  const tabs = [
    { name: "Referrers", resource: "top-referrers", searchParam: "referrer" },
    { name: "Sources", resource: "top-utm-sources", searchParam: "utm-source" },
    { name: "Mediums", resource: "top-utm-mediums", searchParam: "utm-medium" },
    {
      name: "Campaigns",
      resource: "top-utm-campaigns",
      searchParam: "utm-campaign",
    },
  ];

  return (
    <Card
      class="min-h-64"
      size="big"
    >
      <Tabs
        tabs={tabs.map(
          ({ name, resource, searchParam }, i) => ({
            name,
            children: (
              <TopBarGauge
                key={name}
                resource={resource}
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
