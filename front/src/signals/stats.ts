import { computed, effect, Signal, signal } from "@preact/signals";
import { DataFrame, emptyDataFrame } from "@/lib/types.ts";
import * as location from "@/signals/location.ts";
import { load } from "@/signals/loading.ts";

/**
 * Stat define a data frame loaded from Prisme stats API.
 */
export type Stat<K = number> = DataFrame<K> & { loading: boolean };

const loadingStat = { ...emptyDataFrame, loading: true };

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
  signal: Signal<Stat<K>>,
  stat: string,
  init?: RequestInit,
) => {
  const search = location.current.value.search;
  signal.value = { ...loadingStat };
  load(async () => {
    const r = await fetch(`/api/v1/stats/${stat}${search}`, {
      ...init,
      signal: abort.value.signal,
    });
    const df = await r.json();
    signal.value = { ...df, loading: false };
  });
};

export const visitors = signal({ ...loadingStat });
effect(() => {
  fetchStat(visitors, "visitors");
});
export const totalVisitors = computed(() => sum(visitors.value.values));

export const sessions = signal({ ...loadingStat });
effect(() => fetchStat(sessions, "sessions"));
export const totalSessions = computed(() => sum(sessions.value.values));

export const pageViews = signal({ ...loadingStat });
effect(() => fetchStat(pageViews, "pageviews"));
export const totalPageViews = computed(() => sum(pageViews.value.values));

export const sessionsDuration = signal({ ...loadingStat });
effect(() => fetchStat(sessionsDuration, "sessions-duration"));
export const avgSessionsDuration = computed(() =>
  avg(sessionsDuration.value.values)
);

export const bounces = signal({ ...loadingStat });
effect(() => fetchStat(bounces, "bounces"));

export const liveVisitors = signal({ ...loadingStat });
effect(() => fetchStat(liveVisitors, "live-visitors"));
export const totalLiveVisitors = computed(() => avg(liveVisitors.value.values));

export const viewsPerSessions = computed(() => {
  if (totalPageViews.value > 0 && totalSessions.value > 0) {
    return totalPageViews.value / totalSessions.value;
  }
  return 0;
});

export const topPages = signal<Stat<string>>({ ...loadingStat });
effect(() => fetchStat(topPages, "top-pages"));

export const topEntryPages = signal<Stat<string>>({ ...loadingStat });
effect(() => fetchStat(topEntryPages, "top-entry-pages"));

export const topExitPages = signal<Stat<string>>({ ...loadingStat });
effect(() => fetchStat(topExitPages, "top-exit-pages"));

export const topCountries = signal<Stat<string>>({ ...loadingStat });
effect(() => fetchStat(topCountries, "top-countries"));

export const topBrowsers = signal<Stat<string>>({ ...loadingStat });
effect(() => fetchStat(topBrowsers, "top-browsers"));

export const topOperatingSystems = signal<Stat<string>>({ ...loadingStat });
effect(() => fetchStat(topOperatingSystems, "top-operating-systems"));

export const topReferrers = signal<Stat<string>>({
  ...loadingStat,
});
effect(() => fetchStat(topReferrers, "top-referrers"));

export const topUtmSources = signal<Stat<string>>({
  ...loadingStat,
});
effect(() => fetchStat(topUtmSources, "top-utm-sources"));

export const topUtmMediums = signal<Stat<string>>({
  ...loadingStat,
});
effect(() => fetchStat(topUtmMediums, "top-utm-mediums"));

export const topUtmCampaigns = signal<Stat<string>>({
  ...loadingStat,
});
effect(() => fetchStat(topUtmCampaigns, "top-utm-campaigns"));
