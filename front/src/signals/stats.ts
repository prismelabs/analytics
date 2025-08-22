import { computed, effect, signal } from "@preact/signals";
import { emptyDataFrame } from "@/lib/types.ts";
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

const fetchStat = async (stat: string, init?: RequestInit) => {
  const search = location.current.value.search;
  const r = await fetch(`/api/v1/stats/${stat}${search}`, {
    ...init,
    signal: abort.value.signal,
  });
  return await r.json();
};

export const visitors = signal({ ...emptyDataFrame });
effect(() => {
  fetchStat("visitors").then((df) => visitors.value = df);
});
export const totalVisitors = computed(() => sum(visitors.value.values));

export const sessions = signal({ ...emptyDataFrame });
effect(() => {
  fetchStat("sessions").then((df) => sessions.value = df);
});
export const totalSessions = computed(() => sum(sessions.value.values));

export const pageViews = signal({ ...emptyDataFrame });
effect(() => {
  fetchStat("pageviews").then((df) => pageViews.value = df);
});
export const totalPageViews = computed(() => sum(pageViews.value.values));

export const sessionsDuration = signal({ ...emptyDataFrame });
effect(() => {
  fetchStat("sessions-duration").then((df) => sessionsDuration.value = df);
});
export const avgSessionsDuration = computed(() =>
  avg(sessionsDuration.value.values)
);

export const bounces = signal({ ...emptyDataFrame });
effect(() => {
  fetchStat("bounces").then((df) => bounces.value = df);
});

export const liveVisitors = signal({ ...emptyDataFrame });
effect(() => {
  fetchStat("live-visitors").then((df) => liveVisitors.value = df);
});
export const totalLiveVisitors = computed(() => avg(liveVisitors.value.values));

export const viewsPerSessions = computed(() => {
  if (totalPageViews.value > 0 && totalSessions.value > 0) {
    return totalPageViews.value / totalSessions.value;
  }
  return 0;
});
