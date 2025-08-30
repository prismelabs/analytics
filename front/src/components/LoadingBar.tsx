import { Signal } from "@preact/signals";

export default function (
  { loading, class: className }: {
    loading: Signal<boolean> | boolean;
    class?: string;
  },
) {
  const isLoading = typeof loading === "object" ? loading.value : loading;

  return (
    <span
      class={`loader h-4 top-0 ${isLoading ? "" : "hidden!"} ${
        className ? className : ""
      }`}
    />
  );
}
