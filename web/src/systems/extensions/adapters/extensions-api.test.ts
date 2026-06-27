import { afterEach, describe, expect, it } from "vitest";

import { installFetchStub, matchPath } from "@/test/utils";

import { listCatalogAgents, listCatalogExtensions } from "./extensions-api";

const WORKSPACE_ID_HEADER = "x-rc-workspace-id";

describe("extensions api adapter", () => {
  let restore: (() => void) | null = null;

  afterEach(() => {
    restore?.();
    restore = null;
  });

  it("Should list agents with workspace header from GET /api/catalog/agents", async () => {
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/catalog/agents"),
        status: 200,
        body: {
          agents: [{ name: "my-agent", scope: "workspace", description: "Does things" }],
        },
      },
    ]);
    restore = stub.restore;

    const result = await listCatalogAgents("ws-1");
    expect(result).toHaveLength(1);
    expect(result[0]?.name).toBe("my-agent");
    expect(stub.calls[0]?.headers[WORKSPACE_ID_HEADER]).toBe("ws-1");
  });

  it("Should return empty array when agents list is absent", async () => {
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/catalog/agents"),
        status: 200,
        body: { agents: [] },
      },
    ]);
    restore = stub.restore;

    const result = await listCatalogAgents("ws-1");
    expect(result).toHaveLength(0);
  });

  it("Should list extensions with workspace header from GET /api/catalog/extensions", async () => {
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/catalog/extensions"),
        status: 200,
        body: {
          extensions: [
            {
              name: "my-ext",
              source: "workspace",
              enabled: true,
              description: "An extension",
              version: "1.0.0",
            },
          ],
        },
      },
    ]);
    restore = stub.restore;

    const result = await listCatalogExtensions("ws-1");
    expect(result).toHaveLength(1);
    expect(result[0]?.name).toBe("my-ext");
    expect(stub.calls[0]?.headers[WORKSPACE_ID_HEADER]).toBe("ws-1");
  });

  it("Should render sanitized agent warnings without filesystem paths", async () => {
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/catalog/agents"),
        status: 200,
        body: {
          agents: [
            {
              name: "bad-agent",
              scope: "workspace",
              warnings: ["agent definition file (AGENT.md) is missing"],
            },
          ],
        },
      },
    ]);
    restore = stub.restore;

    const result = await listCatalogAgents("ws-1");
    const warning = result[0]?.warnings?.[0] ?? "";
    expect(warning).not.toMatch(/\/Users\/|\/home\/|\.rc\/agents/);
    expect(warning).toBe("agent definition file (AGENT.md) is missing");
  });

  it("Should throw on non-success from GET /api/catalog/agents", async () => {
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/catalog/agents"),
        status: 500,
        body: { code: "internal_error", message: "catalog unavailable", request_id: "r1" },
      },
    ]);
    restore = stub.restore;

    await expect(listCatalogAgents("ws-1")).rejects.toThrow(/catalog unavailable/);
  });
});
