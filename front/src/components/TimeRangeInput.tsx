import { effect, signal } from "@preact/signals";

import { CalendarIcon } from "@heroicons/react/20/solid";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/react";
import { TimeRange } from "@/lib/timerange.ts";
import * as location from "@/signals/location.ts";

export default function TimeRangeInput() {
  return (
    <Menu as="div" class="relative inline-block">
      <MenuButton
        as="button"
        class="bg-trend-background flex items-center gap-2 w-full justify-center rounded-md px-3 py-2 text-sm shadow-sm outline-none"
      >
        <CalendarIcon class="size-6" strokeWidth={2} />
        <p>{selectedTimeRange.value.text}</p>
      </MenuButton>
      <MenuItems
        transition
        className="absolute right-0 z-10 mt-2 w-56 origin-top-right select-none divide-y divide-trend-page/60 rounded-md bg-trend-background shadow-lg transition focus:outline-none data-[closed]:scale-95 data-[closed]:transform data-[closed]:opacity-0 data-[enter]:duration-100 data-[leave]:duration-75 data-[enter]:ease-out data-[leave]:ease-in"
      >
        {timeRanges.map((tr) => (
          <MenuItem
            className="block px-4 py-2 text-sm outline-none hover:bg-trend-primary/10"
            onClick={() => {
              selectedTimeRange.value = tr;
              location.update((url) => {
                url.searchParams.set("from", selectedTimeRange.value.from);
                url.searchParams.set("to", selectedTimeRange.value.to);
              });
            }}
          >
            <span
              class={tr.text === selectedTimeRange.value.text
                ? "font-bold"
                : ""}
            >
              {tr.text}
            </span>
          </MenuItem>
        ))}
      </MenuItems>
    </Menu>
  );
}

const timeRanges: Array<TimeRange & { text: string }> = [
  { text: "Past hour", from: "now-1h", to: "now" },
  { text: "Today", from: "now-1d", to: "now" },
  { text: "Yesterday", from: "now-2d", to: "now-1d" },
  { text: "Last 24 hours", from: "now-24h", to: "now" },
  { text: "Last 7 days", from: "now-7d", to: "now" },
  { text: "Last 14 days", from: "now-14d", to: "now" },
  { text: "Last 30 days", from: "now-30d", to: "now" },
  { text: "Last 90 days", from: "now-90d", to: "now" },
  { text: "Day before yesterday", from: "now-3d", to: "now-2d" },
  { text: "This day last week", from: "now-8d", to: "now-7d" },
];

const selectedTimeRange = signal({
  text: "Last 7 days",
  from: "now-7d",
  to: "now",
});

effect(() => {
  location.update((url) => {
    const from = url.searchParams.get("from");
    const to = url.searchParams.get("to");

    const range = timeRanges.find((range) =>
      range.from === from?.trim()?.replaceAll(" ", "") &&
      range.to === to?.trim()?.replaceAll(" ", "")
    );
    if (!range) {
      url.searchParams.set("from", selectedTimeRange.value.from);
      url.searchParams.set("to", selectedTimeRange.value.to);
    } else {
      selectedTimeRange.value = range;
    }
  });
});
