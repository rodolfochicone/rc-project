import { afterEach, describe, expect, it } from "vitest";

import { installFetchStub, matchPath } from "@/test/utils";

import { startExec } from "./exec-api";

const startedRunResponse = {
  run: {
    run_id: "run-exec",
    mode: "exec",
    presentation_mode: "detach",
    workspace_id: "ws-1",
    started_at: "2026-01-01T00:00:00Z",
    status: "queued",
  },
};

describe("exec api adapter", () => {
  let restore: (() => void) | null = null;

  afterEach(() => {
    restore?.();
    restore = null;
  });

  it("Should include interactive: true in the body only when the flag is set", async () => {
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/exec", "POST"),
        status: 201,
        body: startedRunResponse,
      },
    ]);
    restore = stub.restore;

    await startExec({ workspacePath: "/tmp/ws", prompt: "hi", interactive: true });

    const body = JSON.parse(stub.calls[0]?.body ?? "{}");
    expect(body.interactive).toBe(true);
  });

  it("Should omit interactive when the flag is not set", async () => {
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/exec", "POST"),
        status: 201,
        body: startedRunResponse,
      },
    ]);
    restore = stub.restore;

    await startExec({ workspacePath: "/tmp/ws", prompt: "hi" });

    const body = JSON.parse(stub.calls[0]?.body ?? "{}");
    expect(body.interactive).toBeUndefined();
  });
});
