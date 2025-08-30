import { effect, signal } from "@preact/signals";

export const current = signal(new URL(globalThis.location.toString()));

export const update = (update: (_: URL) => void) => {
  const loc = new URL(current.value.toString());
  update(loc);
  if (loc.toString() !== current.value.toString()) current.value = loc;
};

// Sync location signals with browser URL bar.
effect(() => {
  if (current.value.toString() !== globalThis.location.toString()) {
    history.replaceState(null, "", current.value.toString());
  }
});
