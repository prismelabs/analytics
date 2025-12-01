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
import Filters from "@/components/Filters.tsx";
import LoadingBar from "@/components/LoadingBar.tsx";
import Logo from "@/components/Logo.tsx";

export function App() {
  const _ = trend.value;

  return (
    <>
      <LoadingBar loading={isLoading} class="fixed! top-0 z-50 h-2" />
      <header class="px-4 py-2 w-full flex gap-2 justify-between sticky bg-trend-page top-0 z-40">
        <div class="flex items-center gap-2 min-w-max">
          <div class="bg-black dark:bg-white">
            <Logo class="size-8 mix-blend-difference dark:saturate-200" />
          </div>
          <span class="text-system-fg font-bold">Prisme Analytics</span>
        </div>
        <TimeRangeInput />
      </header>
      <main class="p-4 md:px-16 lg:px-32 m-auto flex flex-col gap-4">
        <Filters />
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
