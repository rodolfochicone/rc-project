import { afterEach, describe, expect, it } from "vitest";

import { installFetchStub, matchPath } from "@/test/utils";

import {
  getGlobalConfig,
  getWorkspaceConfig,
  putGlobalConfig,
  putWorkspaceConfig,
} from "./config-api";

const WORKSPACE_ID_HEADER = "x-rc-workspace-id";

describe("config api adapter", () => {
  let restore: (() => void) | null = null;

  afterEach(() => {
    restore?.();
    restore = null;
  });

  it("Should return config from GET /api/config/global", async () => {
    const keepMax = 42;
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/config/global"),
        status: 200,
        body: { config: { runs: { keep_max: keepMax } } },
      },
    ]);
    restore = stub.restore;

    const result = await getGlobalConfig();
    expect(result.runs?.keep_max).toBe(keepMax);
    expect(stub.calls[0]?.method).toBe("GET");
  });

  it("Should throw on non-success from GET /api/config/global", async () => {
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/config/global"),
        status: 500,
        body: { code: "internal_error", message: "daemon error", request_id: "r1" },
      },
    ]);
    restore = stub.restore;

    await expect(getGlobalConfig()).rejects.toThrow(/daemon error/);
  });

  it("Should PUT config to /api/config/global and return updated doc", async () => {
    const keepMax = 7;
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/config/global", "PUT"),
        status: 200,
        body: { config: { runs: { keep_max: keepMax } } },
      },
    ]);
    restore = stub.restore;

    const result = await putGlobalConfig({ runs: { keep_max: keepMax } });
    expect(result.runs?.keep_max).toBe(keepMax);
    expect(stub.calls[0]?.method).toBe("PUT");
  });

  it("Should throw with config_invalid message on 400 from PUT /api/config/global", async () => {
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/config/global", "PUT"),
        status: 400,
        body: { code: "config_invalid", message: "provider must not be empty", request_id: "r2" },
      },
    ]);
    restore = stub.restore;

    await expect(putGlobalConfig({ fetch_reviews: { provider: "" } })).rejects.toThrow(
      /provider must not be empty/
    );
  });

  it("Should send X-rc-Workspace-ID header for GET /api/config/workspace", async () => {
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/config/workspace"),
        status: 200,
        body: { config: {} },
      },
    ]);
    restore = stub.restore;

    await getWorkspaceConfig("ws-1");
    expect(stub.calls[0]?.headers[WORKSPACE_ID_HEADER]).toBe("ws-1");
  });

  it("Should send X-rc-Workspace-ID header for PUT /api/config/workspace", async () => {
    const keepMax = 3;
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/config/workspace", "PUT"),
        status: 200,
        body: { config: { runs: { keep_max: keepMax } } },
      },
    ]);
    restore = stub.restore;

    const result = await putWorkspaceConfig("ws-1", { runs: { keep_max: keepMax } });
    expect(result.runs?.keep_max).toBe(keepMax);
    expect(stub.calls[0]?.headers[WORKSPACE_ID_HEADER]).toBe("ws-1");
  });
});
