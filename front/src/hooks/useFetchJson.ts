import { Inputs, useEffect } from "preact/hooks";
import { load } from "@/signals/loading.ts";

export default function <T>(
  info: RequestInfo | URL,
  inputs: Inputs,
  init?: RequestInit,
): Promise<T> {
  return new Promise((res, rej) => {
    useEffect(() => {
      const abort = new AbortController();

      load(() =>
        fetch(info, {
          ...init,
          signal: abort.signal,
        }).then((r) => r.json())
          .then(res)
          .catch(rej)
      );

      return () => abort.abort();
    }, inputs);
  });
}
