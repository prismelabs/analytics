import { computed } from "@preact/signals";
import * as location from "@/signals/location.ts";

export const from = computed(() => {
  const from = location.current.value.searchParams.get("from");
  if (!from) {
    location.update((url) => {
      url.searchParams.set("from", "now-7d");
      url.searchParams.set("to", "now");
    });
    return "";
  }

  return from;
});

export const to = computed(() => {
  const to = location.current.value.searchParams.get("to");
  if (!to) {
    location.update((url) => {
      url.searchParams.set("from", "now-7d");
      url.searchParams.set("to", "now");
    });
    return "";
  }

  return to;
});
