import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/react";
import { ChevronDownIcon } from "@heroicons/react/20/solid";

export default function Dropdown() {
  return (
    <Menu as="div" class="relative inline-block">
      <MenuButton class="inline-flex w-full justify-center gap-x-1.5 rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50">
        Options
        <ChevronDownIcon
          aria-hidden="true"
          class="-mr-1 size-5 text-gray-400"
        />
      </MenuButton>

      <MenuItems
        transition
        class="absolute right-0 z-10 mt-2 w-56 origin-top-right divide-y divide-gray-100 rounded-md bg-white shadow-lg ring-1 ring-black/5 transition focus:outline-none data-[closed]:scale-95 data-[closed]:transform data-[closed]:opacity-0 data-[enter]:duration-100 data-[leave]:duration-75 data-[enter]:ease-out data-[leave]:ease-in"
      >
        <div class="py-1">
          <MenuItem>
            <a
              href="#"
              class="block px-4 py-2 text-sm text-gray-700 data-[focus]:bg-gray-100 data-[focus]:text-gray-900 data-[focus]:outline-none"
            >
              Edit
            </a>
          </MenuItem>
          <MenuItem>
            <a
              href="#"
              class="block px-4 py-2 text-sm text-gray-700 data-[focus]:bg-gray-100 data-[focus]:text-gray-900 data-[focus]:outline-none"
            >
              Duplicate
            </a>
          </MenuItem>
        </div>
        <div class="py-1">
          <MenuItem>
            <a
              href="#"
              class="block px-4 py-2 text-sm text-gray-700 data-[focus]:bg-gray-100 data-[focus]:text-gray-900 data-[focus]:outline-none"
            >
              Archive
            </a>
          </MenuItem>
          <MenuItem>
            <a
              href="#"
              class="block px-4 py-2 text-sm text-gray-700 data-[focus]:bg-gray-100 data-[focus]:text-gray-900 data-[focus]:outline-none"
            >
              Move
            </a>
          </MenuItem>
        </div>
        <div class="py-1">
          <MenuItem>
            <a
              href="#"
              class="block px-4 py-2 text-sm text-gray-700 data-[focus]:bg-gray-100 data-[focus]:text-gray-900 data-[focus]:outline-none"
            >
              Share
            </a>
          </MenuItem>
          <MenuItem>
            <a
              href="#"
              class="block px-4 py-2 text-sm text-gray-700 data-[focus]:bg-gray-100 data-[focus]:text-gray-900 data-[focus]:outline-none"
            >
              Add to favorites
            </a>
          </MenuItem>
        </div>
        <div class="py-1">
          <MenuItem>
            <a
              href="#"
              class="block px-4 py-2 text-sm text-gray-700 data-[focus]:bg-gray-100 data-[focus]:text-gray-900 data-[focus]:outline-none"
            >
              Delete
            </a>
          </MenuItem>
        </div>
      </MenuItems>
    </Menu>
  );
}
