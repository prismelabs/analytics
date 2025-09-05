import { computed, effect, Signal, signal } from "@preact/signals";
import { DataFrame, emptyDataFrame, error, ok, Result } from "@/lib/types.ts";
import * as location from "@/signals/location.ts";
import { load } from "@/signals/loading.ts";

/**
 * Stat defines a data frame loaded from Prisme stats API.
 */
export type Stat<K = number> = DataFrame<K> & { loading: boolean };

/**
 * FetchedStat defines result of a Stat<K> fetched from Prisme.
 */
export type FetchedStat<K = number> = Result<Stat<K>, string>;

const loadingStat = { ...emptyDataFrame, loading: true };

const sum = (arr?: Array<number>) =>
  arr ? arr.reduce((acc, v) => acc + v, 0) : 0;
const avg = (arr?: Array<number>) =>
  arr ? arr.length === 0 ? 0 : sum(arr) / arr.length : 0;

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
  signal: Signal<FetchedStat<K>>,
  stat: string,
  init?: RequestInit,
) => {
  const search = location.current.value.search;
  signal.value = { ok: { ...loadingStat }, error: null };
  load(async () => {
    try {
      const r = await fetch(`/api/v1/stats/${stat}${search}`, {
        ...init,
        signal: abort.value.signal,
      });
      const df = await r.json();
      signal.value = { ok: { ...df, loading: false }, error: null };
    } catch (err) {
      console.error(err);
      signal.value = error("An unexpected error occured");
    }
  });
};

export const visitors = signal(ok({ ...loadingStat }));
effect(() => {
  fetchStat(visitors, "visitors");
});
export const totalVisitors = computed(() => sum(visitors.value.ok?.values));

export const sessions = signal(ok({ ...loadingStat }));
effect(() => fetchStat(sessions, "sessions"));
export const totalSessions = computed(() => sum(sessions.value.ok?.values));

export const pageViews = signal(ok({ ...loadingStat }));
effect(() => fetchStat(pageViews, "pageviews"));
export const totalPageViews = computed(() => sum(pageViews.value.ok?.values));

export const sessionsDuration = signal(ok({ ...loadingStat }));
effect(() => fetchStat(sessionsDuration, "sessions-duration"));
export const avgSessionsDuration = computed(() =>
  avg(sessionsDuration.value.ok?.values)
);

export const bounces = signal(ok({ ...loadingStat }));
effect(() => fetchStat(bounces, "bounces"));

export const liveVisitors = signal(ok({ ...loadingStat }));
effect(() => fetchStat(liveVisitors, "live-visitors"));
export const totalLiveVisitors = computed(() =>
  avg(liveVisitors.value.ok?.values)
);

export const viewsPerSessions = computed(() => {
  if (totalPageViews.value > 0 && totalSessions.value > 0) {
    return totalPageViews.value / totalSessions.value;
  }
  return 0;
});

export const topPages = signal<FetchedStat<string>>(ok({ ...loadingStat }));
effect(() => fetchStat(topPages, "top-pages"));

export const topEntryPages = signal<FetchedStat<string>>(
  ok({ ...loadingStat }),
);
effect(() => fetchStat(topEntryPages, "top-entry-pages"));

export const topExitPages = signal<FetchedStat<string>>(ok({ ...loadingStat }));
effect(() => fetchStat(topExitPages, "top-exit-pages"));

export const topCountries = signal<FetchedStat<string>>(ok({ ...loadingStat }));
effect(() => fetchStat(topCountries, "top-countries"));

export const topBrowsers = signal<FetchedStat<string>>(ok({ ...loadingStat }));
effect(() => fetchStat(topBrowsers, "top-browsers"));

export const topOperatingSystems = signal<FetchedStat<string>>(ok({
  ...loadingStat,
}));
effect(() => fetchStat(topOperatingSystems, "top-operating-systems"));

export const topReferrers = signal<FetchedStat<string>>(ok({
  ...loadingStat,
}));
effect(() => fetchStat(topReferrers, "top-referrers"));

export const topUtmSources = signal<FetchedStat<string>>(ok({
  ...loadingStat,
}));
effect(() => fetchStat(topUtmSources, "top-utm-sources"));

export const topUtmMediums = signal<FetchedStat<string>>(ok({
  ...loadingStat,
}));
effect(() => fetchStat(topUtmMediums, "top-utm-mediums"));

export const topUtmCampaigns = signal<FetchedStat<string>>(ok({
  ...loadingStat,
}));
effect(() => fetchStat(topUtmCampaigns, "top-utm-campaigns"));
