import TopLocations from "@/components/TopLocations.tsx";
import Summary from "@/components/Summary.tsx";
import TimeRangeInput from "@/components/TimeRangeInput.tsx";
import TimeSerie from "@/components/TimeSerie.tsx";
import TopPages from "@/components/TopPages.tsx";
import TopBrowsers from "@/components/TopBrowsers.tsx";
import TopOperatingSystems from "@/components/TopOperatingSystems.tsx";
import TopSources from "@/components/TopSources.tsx";
import { isLoading } from "@/signals/loading.ts";
import { trend } from "@/signals/trend.ts";

export function App() {
  const _ = trend.value;

  return (
    <>
      <span class={`loader fixed top-0 ${isLoading.value ? "" : "hidden!"}`}>
      </span>
      <header class="p-4 flex justify-between">
        <div />
        <TimeRangeInput />
      </header>
      <main class="p-4 md:px-16 lg:px-32 m-auto flex flex-col gap-4">
        <Summary />
        <TimeSerie />
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <TopPages />
          <TopSources />
        </div>
        <TopLocations />
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <TopBrowsers />
          <TopOperatingSystems />
        </div>
      </main>
    </>
  );
}
