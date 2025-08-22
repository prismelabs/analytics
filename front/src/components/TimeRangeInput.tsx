import { computed, effect, signal } from "@preact/signals";
import { CalendarDateRangeIcon } from "@heroicons/react/20/solid";
import { Popover, PopoverButton, PopoverPanel } from "@headlessui/react";
import Calendar from "react-calendar";
import { TimeRange } from "@/lib/timerange.ts";
import * as location from "@/signals/location.ts";

export default function TimeRangeInput() {
  return (
    <Popover as="div" class="relative inline-block text-system-fg">
      <PopoverButton
        as="button"
        class="bg-trend-background flex items-center gap-2 w-full justify-center rounded-xl p-3 text-sm outline-none"
      >
        <CalendarDateRangeIcon class="size-6" strokeWidth={2} />
        <p>{selectedTimeRangeText.value}</p>
      </PopoverButton>
      <PopoverPanel
        transition
        className="absolute right-0 z-10 mt-2 w-64 h-128 origin-top-right select-none divide-y divide-trend-page/60 rounded-md bg-trend-background shadow-lg transition focus:outline-none data-[closed]:scale-95 data-[closed]:transform data-[closed]:opacity-0 data-[enter]:duration-100 data-[leave]:duration-75 data-[enter]:ease-out data-[leave]:ease-in overflow-hidden flex flex-col pl-2"
      >
        {({ close }: { close: () => void }) => (
          <>
            <div class="text-black p-2">
              <Calendar
                selectRange
                onChange={([from, to]: [Date, Date]) => {
                  selectedTimeRange.value = {
                    from: from.toISOString(),
                    to: to.toISOString(),
                  };
                  close();
                }}
              />
            </div>
            <ul class="overflow-y-auto h-full py-2">
              {timeRanges.map((tr) => (
                <span
                  class={`block text-sm outline-none p-2 mr-2 rounded hover:bg-trend-primary-light ${
                    tr.text === selectedTimeRangeText.value ? "font-bold" : ""
                  }`}
                  onClick={() => {
                    selectedTimeRange.value = tr;
                    close();
                  }}
                >
                  {tr.text}
                </span>
              ))}
            </ul>
          </>
        )}
      </PopoverPanel>
    </Popover>
  );
}

const timeRanges: Array<TimeRange & { text: string }> = [
  { text: "Last 5 minutes", from: "now-5m", to: "now" },
  { text: "Last 15 minutes", from: "now-15m", to: "now" },
  { text: "Last 30 minutes", from: "now-30m", to: "now" },
  { text: "Last 1 hour", from: "now-1h", to: "now" },
  { text: "Last 3 hours", from: "now-3h", to: "now" },
  { text: "Last 6 hours", from: "now-6h", to: "now" },
  { text: "Last 12 hours", from: "now-12h", to: "now" },
  { text: "Last 24 hours", from: "now-24h", to: "now" },
  { text: "Last 2 days", from: "now-2d", to: "now" },
  { text: "Last 7 days", from: "now-7d", to: "now" },
  { text: "Last 30 days", from: "now-30d", to: "now" },
  { text: "Last 90 days", from: "now-90d", to: "now" },
  { text: "Last 6 months", from: "now-6M", to: "now" },
  { text: "Last 1 year", from: "now-1y", to: "now" },
  { text: "Last 2 year", from: "now-2y", to: "now" },
  { text: "Last 5 year", from: "now-5y", to: "now" },
  { text: "Yesterday", from: "now-1d/d", to: "now-1d/d" },
  { text: "Day before yesterday", from: "now-2d/d", to: "now-2d/d" },
  { text: "This day last week", from: "now-7d/d", to: "now-7d/d" },
  { text: "Previous week", from: "now-1w/w", to: "now-1w/w" },
  { text: "Previous month", from: "now-1M/M", to: "now-1M/M" },
  { text: "Previous fiscal quarter", from: "now-1Q/fQ", to: "now-1Q/fQ" },
  { text: "Previous year", from: "now-1y/y", to: "now-1y/y" },
  { text: "Today", from: "now/d", to: "now/d" },
  { text: "Today so far", from: "now/d", to: "now" },
  { text: "This week", from: "now/w", to: "now/w" },
  { text: "This week so far", from: "now/w", to: "now" },
  { text: "This month", from: "now/M", to: "now/M" },
  { text: "This month so far", from: "now/M", to: "now" },
  { text: "This year", from: "now/y", to: "now/y" },
  { text: "This year so far", from: "now/y", to: "now" },
  { text: "This fiscal quarter", from: "now/fQ", to: "now/fQ" },
  { text: "This fiscal quarter so far", from: "now/fQ", to: "now" },
];

export const selectedTimeRange = signal<TimeRange>(timeRanges[9]);

const selectedTimeRangeText = computed(() => {
  const range = timeRanges.find((range) =>
    range.from === selectedTimeRange.value.from?.trim()?.replaceAll(" ", "") &&
    range.to === selectedTimeRange.value.to?.trim()?.replaceAll(" ", "")
  );

  if (range) return range.text;

  const from = new Date(selectedTimeRange.value.from);
  const to = new Date(selectedTimeRange.value.to);
  const sameD = from.getFullYear() == to.getFullYear() &&
    from.getMonth() == to.getMonth() && from.getDate() == to.getDate();

  return sameD
    ? `${from.toLocaleTimeString()} - ${to.toLocaleTimeString()}`
    : `${from.toLocaleDateString()} - ${to.toLocaleDateString()}`;
});

// Sync selectedTimeRange with URL.
effect(() => {
  const url = location.current.value;
  const from = url.searchParams.get("from");
  const to = url.searchParams.get("to");

  const syncUrl = () =>
    location.update((url) => {
      url.searchParams.set("from", selectedTimeRange.value.from);
      url.searchParams.set("to", selectedTimeRange.value.to);
    });

  if (!from || !to) return syncUrl();

  // Lookup for predefined range.
  const range = timeRanges.find((range) =>
    range.from === from?.trim()?.replaceAll(" ", "") &&
    range.to === to?.trim()?.replaceAll(" ", "")
  );

  // No predefined range found.
  if (!range) {
    // Try to parse date.
    const fromTs = Date.parse(from);
    const toTs = Date.parse(to);

    // Valid date.
    if (!Number.isNaN(fromTs) && !Number.isNaN(toTs)) {
      selectedTimeRange.value = {
        from: from,
        to: to,
      };
      return;
    }

    // Invalid date.
    return syncUrl();
  }

  // Predefined range found.
  selectedTimeRange.value = range;
});

// Sync URL with selectedTimeRange.
effect(() => {
  location.update((url) => {
    url.searchParams.set(
      "from",
      selectedTimeRange.value.from,
    );
    url.searchParams.set("to", selectedTimeRange.value.to);
  });
});
