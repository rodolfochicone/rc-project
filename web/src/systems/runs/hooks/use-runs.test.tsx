import { renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import {
  createTestQueryClient,
  flushAsync,
  installFetchStub,
  matchPath,
  withQuery,
} from "@/test/utils";

import { runKeys } from "../lib/query-keys";
import { useSendRunInput } from "./use-runs";

describe("useSendRunInput", () => {
  let restore: (() => void) | null = null;

  afterEach(() => {
    restore?.();
    restore = null;
  });

  it("Should invalidate the run, snapshot, and transcript queries on success", async () => {
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/runs/run-7/input", "POST"),
        status: 202,
        body: { accepted: true },
      },
    ]);
    restore = stub.restore;

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

    const { result } = renderHook(() => useSendRunInput(), {
      wrapper: withQuery(queryClient),
    });

    await result.current.mutateAsync({
      runId: "run-7",
      input: { prompt_id: "p1", option_id: "a" },
    });
    await flushAsync();

    const invalidatedKeys = invalidateSpy.mock.calls.map(([arg]) => JSON.stringify(arg?.queryKey));
    expect(invalidatedKeys).toContain(JSON.stringify(runKeys.run("run-7")));
    expect(invalidatedKeys).toContain(JSON.stringify(runKeys.snapshot("run-7")));
    expect(invalidatedKeys).toContain(JSON.stringify(runKeys.transcript("run-7")));
  });
});
