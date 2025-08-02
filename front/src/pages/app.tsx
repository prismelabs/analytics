import Summary from "@/components/Summary.tsx";
import TimeSerie from "@/components/TimeSerie.tsx";
import BarGauge from "@/components/BarGauge.tsx";
import Locations from "@/components/Locations.tsx";
import Card from "@/components/Card.tsx";
import TimeRangeInput from "@/components/TimeRangeInput.tsx";
import Tabs from "@/components/Tabs.tsx";

export function App() {
  return (
    <>
      <header class="p-4 flex justify-between">
        <div />
        <TimeRangeInput />
      </header>
      <main class="p-4 md:px-16 lg:px-32 m-auto flex flex-col gap-4">
        <Summary />
        <TimeSerie />
        <section class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Card
            class="h-64"
            size="big"
          >
            <Tabs
              tabs={[{
                name: "Pages",
                children: (
                  <BarGauge
                    data={[
                      { label: "Foo", value: 100 },
                      { label: "Bar", value: 50 },
                      { label: "Baz", value: 0 },
                    ]}
                  />
                ),
              }, {
                name: "Entry Pages",
                children: (
                  <BarGauge
                    data={[
                      { label: "Foo", value: 0 },
                      { label: "Bar", value: 50 },
                      { label: "Baz", value: 100 },
                    ]}
                  />
                ),
              }]}
            >
            </Tabs>
          </Card>
          <Card
            title="Top pages"
            class="h-64"
            size="big"
          >
            <BarGauge
              data={[
                { label: "www.prismeanalytics.com", value: 100 },
                { label: "Bar", value: 50 },
                { label: "Baz", value: 0 },
              ]}
            />
          </Card>
        </section>
        <Locations />
      </main>
    </>
  );
}
