import { effect, signal } from "@preact/signals";

export const current = signal(new URL(globalThis.location.toString()));

export const update = (update: (_: URL) => void) => {
  const loc = new URL(current.value.toString());
  update(loc);
  loc.searchParams.forEach((v, k) => {
    if (v.trim() === "") loc.searchParams.delete(k);
  });
  if (loc.toString() !== current.value.toString()) current.value = loc;
};

// Sync location signals with browser URL bar.
effect(() => {
  if (current.value.toString() !== globalThis.location.toString()) {
    history.replaceState(null, "", current.value.toString());
  }
});
