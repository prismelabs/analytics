import { computed, effect, Signal, signal } from "@preact/signals";
import { DataFrame, emptyDataFrame } from "@/lib/types.ts";
import * as location from "@/signals/location.ts";

const sum = (arr: Array<number>) => arr.reduce((acc, v) => acc + v, 0);
const avg = (arr: Array<number>) =>
  arr.length === 0 ? 0 : sum(arr) / arr.length;

// We update abort signal every time the URL changes.
const abort = computed(() => {
  // Depend on URL.
  location.current.value;
  return new AbortController();
});

// Abort previous controller.
(() => {
  let controller: AbortController | null = null;
  effect(() => {
    if (controller) controller.abort();
    controller = abort.value;
  });
})();

const fetchStat = <K>(
  signal: Signal<DataFrame<K>>,
  stat: string,
  init?: RequestInit,
) => {
  const search = location.current.value.search;
  fetch(`/api/v1/stats/${stat}${search}`, {
    ...init,
    signal: abort.value.signal,
  }).then((r) => r.json())
    .then((df) => signal.value = df);
};

export const visitors = signal<DataFrame>({ ...emptyDataFrame });
effect(() => {
  fetchStat(visitors, "visitors");
});
export const totalVisitors = computed(() => sum(visitors.value.values));

export const sessions = signal<DataFrame>({ ...emptyDataFrame });
effect(() => fetchStat(sessions, "sessions"));
export const totalSessions = computed(() => sum(sessions.value.values));

export const pageViews = signal<DataFrame>({ ...emptyDataFrame });
effect(() => fetchStat(pageViews, "pageviews"));
export const totalPageViews = computed(() => sum(pageViews.value.values));

export const sessionsDuration = signal<DataFrame>({ ...emptyDataFrame });
effect(() => fetchStat(sessionsDuration, "sessions-duration"));
export const avgSessionsDuration = computed(() =>
  avg(sessionsDuration.value.values)
);

export const bounces = signal<DataFrame>({ ...emptyDataFrame });
effect(() => fetchStat(bounces, "bounces"));

export const liveVisitors = signal<DataFrame>({ ...emptyDataFrame });
effect(() => fetchStat(liveVisitors, "live-visitors"));
export const totalLiveVisitors = computed(() => avg(liveVisitors.value.values));

export const viewsPerSessions = computed(() => {
  if (totalPageViews.value > 0 && totalSessions.value > 0) {
    return totalPageViews.value / totalSessions.value;
  }
  return 0;
});

export const topPages = signal<DataFrame<string>>({ ...emptyDataFrame });
effect(() => fetchStat(topPages, "top-pages"));

export const topEntryPages = signal<DataFrame<string>>({ ...emptyDataFrame });
effect(() => fetchStat(topEntryPages, "top-entry-pages"));

export const topExitPages = signal<DataFrame<string>>({ ...emptyDataFrame });
effect(() => fetchStat(topExitPages, "top-exit-pages"));

export const topCountries = signal<DataFrame<string>>({ ...emptyDataFrame });
effect(() => fetchStat(topCountries, "top-countries"));

export const topBrowsers = signal<DataFrame<string>>({ ...emptyDataFrame });
effect(() => fetchStat(topBrowsers, "top-browsers"));

export const topOperatingSystems = signal<DataFrame<string>>({
  ...emptyDataFrame,
});
effect(() => fetchStat(topOperatingSystems, "top-operating-systems"));

export const topReferrers = signal<DataFrame<string>>({
  ...emptyDataFrame,
});
effect(() => fetchStat(topReferrers, "top-referrers"));

export const topUtmSources = signal<DataFrame<string>>({
  ...emptyDataFrame,
});
effect(() => fetchStat(topUtmSources, "top-utm-sources"));

export const topUtmMediums = signal<DataFrame<string>>({
  ...emptyDataFrame,
});
effect(() => fetchStat(topUtmMediums, "top-utm-mediums"));

export const topUtmCampaigns = signal<DataFrame<string>>({
  ...emptyDataFrame,
});
effect(() => fetchStat(topUtmCampaigns, "top-utm-campaigns"));
