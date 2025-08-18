import { computed, signal } from "@preact/signals";

const loadingCount = signal(0);

export function load<T>(loader: () => Promise<T>): Promise<T> {
  loadingCount.value += 1;
  return loader().finally(() => loadingCount.value -= 1);
}

export const isLoading = computed(() => loadingCount.value !== 0);
