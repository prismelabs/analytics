import { signal, useSignal } from "@preact/signals";

import { CalendarIcon } from "@heroicons/react/20/solid";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/react";
import { TimeRange } from "@/lib/timerange.ts";

export const from = signal("now-24h");
export const to = signal("now");
const label = signal("Last 24 hours");

export default function TimeRangeInput() {
  const timeRanges: Array<TimeRange & { text: string }> = [
    { text: "Past hour", from: "now-1h", to: "now" },
    { text: "Today", from: "now-1d", to: "now" },
    { text: "Yesterday", from: "now-2d", to: "now-1d" },
    { text: "Last 24 hours", from: "now-24h", to: "now" },
    { text: "Last 7 days", from: "now-7d", to: "now" },
    { text: "Last 14 days", from: "now-14d", to: "now" },
    { text: "Last 30 days", from: "now-30d", to: "now" },
    { text: "Last 90 days", from: "now-30d", to: "now" },
    { text: "Day before yesterday", from: "now-3d", to: "now-2d" },
    { text: "This day last week", from: "now-8d", to: "now-7d" },
  ];

  return (
    <Menu as="div" class="relative inline-block">
      <MenuButton
        as="button"
        class="bg-system-bg flex items-center gap-2 w-full justify-center rounded-md px-3 py-2 text-sm shadow-sm outline-none"
      >
        <CalendarIcon class="size-6" strokeWidth={2} />
        <p>{label.value}</p>
      </MenuButton>
      <MenuItems
        transition
        className="absolute right-0 z-10 mt-2 w-56 origin-top-right select-none divide-y divide-system-bg rounded-md bg-system-bg shadow-lg transition focus:outline-none data-[closed]:scale-95 data-[closed]:transform data-[closed]:opacity-0 data-[enter]:duration-100 data-[leave]:duration-75 data-[enter]:ease-out data-[leave]:ease-in"
      >
        {timeRanges.map((tr) => (
          <MenuItem
            className="block px-4 py-2 text-sm outline-none hover:bg-page-bg/50"
            onClick={() => {
              from.value = tr.from;
              to.value = tr.to;
              label.value = tr.text;
            }}
          >
            <span>{tr.text}</span>
          </MenuItem>
        ))}
      </MenuItems>
    </Menu>
  );
}
