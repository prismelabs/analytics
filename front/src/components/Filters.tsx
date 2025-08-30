import * as location from "@/signals/location.ts";
import { MagnifyingGlassIcon, XMarkIcon } from "@heroicons/react/20/solid";
import { countryName } from "@/lib/countries.ts";

export default function () {
  const params = [...location.current.value.searchParams.entries()]
    .filter(([param, values]) =>
      !!filterName[param] && values.trim().length > 0
    );

  if (params.length === 0) return;

  return (
    <div class="bg-trend-primary-light p-2 flex gap-2 items-center rounded-xl sticky top-16 z-30 justify-between items-center">
      <MagnifyingGlassIcon class="size-5" />
      <ul class="flex flex-1 text-trend-foreground text-sm gap-2 items-center overflow-x-auto scrollbar-thin">
        {params.map(([param, value]) => (
          <li class="bg-trend-background rounded-xl py-2 pl-2 pr-1 min-w-max flex">
            {filterName[param]} {value.includes(",") ? "are" : "is"}{" "}
            {value.split(",")
              .map((s) => s.trim())
              .map(mapValue[param] ?? unchanged)
              .join(", ")}
            <XMarkIcon
              class="size-5 cursor-pointer"
              onClick={() =>
                location.update((url) => url.searchParams.delete(param))}
            />
          </li>
        ))}
      </ul>
      <div />
      <XMarkIcon
        class="size-6 cursor-pointer"
        onClick={() =>
          location.update((url) => {
            for (const param in filterName) {
              url.searchParams.delete(param);
            }
          })}
      />
    </div>
  );
}

const filterName: Record<string, string> = {
  "path": "Page",
  "entry-path": "Entry page",
  "exit-path": "Exit page",
  "referrer": "Referrer",
  "os": "OS",
  "browser": "Browser",
  "country": "Country",
  "utm-source": "UTM source",
  "utm-medium": "UTM medium",
  "utm-campaign": "UTM campaign",
};

const mapValue: Record<string, (_: string) => string> = {
  "country": (v: string) => countryName[v] ?? v,
};

function unchanged(v: string) {
  return v;
}
