import { useState } from "preact/hooks";
import { DataFrame, emptyDataFrame } from "@/lib/types.ts";
import * as location from "@/signals/location.ts";
import useFetchJson from "./useFetchJson.ts";

export default function useStats() {
  const [visitors, setVisitors] = useState<DataFrame>(emptyDataFrame);
  const [sessions, setSessions] = useState<DataFrame>(emptyDataFrame);
  const [sessionsDuration, setSessionsDuration] = useState<DataFrame>(
    emptyDataFrame,
  );
  const [pageViews, setPageViews] = useState<DataFrame>(emptyDataFrame);
  const [liveVisitors, setLiveVisitors] = useState<DataFrame>(
    emptyDataFrame,
  );
  const [bounces, setBounces] = useState<DataFrame>(emptyDataFrame);

  // @ts-ignore: vitejs magic env.
  const prismeUrl = import.meta.env.VITE_PRISME_URL;
  const search = location.current.value.search;
  const loc = location.current.value.toString();
  const useFetchStats = (stat: string) =>
    useFetchJson<DataFrame>(`${prismeUrl}/api/v1/stats/${stat}${search}`, [
      loc,
    ]);

  useFetchStats("visitors").then(setVisitors);
  useFetchStats("sessions").then(setSessions);
  useFetchStats("sessions-duration").then(setSessionsDuration);
  useFetchStats("pageviews").then(setPageViews);
  useFetchStats("live-visitors").then(setLiveVisitors);
  useFetchStats("bounces").then(setBounces);

  return {
    visitors,
    sessions,
    sessionsDuration,
    pageViews,
    liveVisitors,
    bounces,
  };
}
