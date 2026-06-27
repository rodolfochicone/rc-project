import { afterEach, describe, expect, it } from "vitest";

import { CSRF_COOKIE_NAME, CSRF_HEADER_NAME } from "@/lib/csrf";
import { installFetchStub, matchPath } from "@/test/utils";

import {
  listWorkspaces,
  registerWorkspace,
  renameWorkspace,
  unregisterWorkspace,
} from "./workspaces-api";

function setCsrfCookie(value: string): void {
  document.cookie = `${CSRF_COOKIE_NAME}=${value}; path=/`;
}

function clearCsrfCookie(): void {
  document.cookie = `${CSRF_COOKIE_NAME}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/`;
}

const workspaceFixture = {
  id: "ws-1",
  name: "My Workspace",
  root_dir: "/tmp/my-ws",
  filesystem_state: "present" as const,
  read_only: false,
  has_catalog_data: true,
  workflow_count: 1,
  run_count: 0,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

describe("workspaces api adapter", () => {
  let restore: (() => void) | null = null;

  afterEach(() => {
    restore?.();
    restore = null;
    clearCsrfCookie();
  });

  it("Should list workspaces from GET /api/workspaces", async () => {
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/workspaces"),
        status: 200,
        body: { workspaces: [workspaceFixture] },
      },
    ]);
    restore = stub.restore;

    const result = await listWorkspaces();
    expect(result).toHaveLength(1);
    expect(result[0]?.id).toBe("ws-1");
    expect(stub.calls[0]?.method).toBe("GET");
  });

  it("Should register a workspace via POST /api/workspaces", async () => {
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/workspaces", "POST"),
        status: 200,
        body: { workspace: { ...workspaceFixture, name: "New WS" } },
      },
    ]);
    restore = stub.restore;

    const result = await registerWorkspace({ name: "New WS", rootDir: "/tmp/new-ws" });
    expect(result.name).toBe("New WS");
    expect(stub.calls[0]?.method).toBe("POST");
    // The daemon contract takes `path` (not `root_dir`); sending the wrong field
    // makes it reject the request with "path is required".
    expect(JSON.parse(stub.calls[0]?.body ?? "{}")).toEqual({
      path: "/tmp/new-ws",
      name: "New WS",
    });
  });

  it("Should throw with server message on register failure", async () => {
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/workspaces", "POST"),
        status: 409,
        body: { code: "conflict", message: "workspace already registered" },
      },
    ]);
    restore = stub.restore;

    await expect(registerWorkspace({ name: "dup", rootDir: "/tmp/dup" })).rejects.toThrow(
      /workspace already registered/
    );
  });

  it("Should rename a workspace via PATCH /api/workspaces/:id", async () => {
    const stub = installFetchStub([
      {
        matcher: (input, init) => {
          const url = typeof input === "string" ? input : (input as Request).url;
          return (
            url.includes("/api/workspaces/ws-1") &&
            ((init?.method ?? "GET").toUpperCase() === "PATCH" ||
              (input instanceof Request && input.method === "PATCH"))
          );
        },
        status: 200,
        body: { workspace: { ...workspaceFixture, name: "Renamed" } },
      },
    ]);
    restore = stub.restore;

    const result = await renameWorkspace({ id: "ws-1", name: "Renamed" });
    expect(result.name).toBe("Renamed");
    expect(stub.calls[0]?.method).toBe("PATCH");
  });

  it("Should unregister a workspace via DELETE /api/workspaces/:id", async () => {
    const stub = installFetchStub([
      {
        matcher: (input, init) => {
          const url = typeof input === "string" ? input : (input as Request).url;
          return (
            url.includes("/api/workspaces/ws-1") &&
            ((init?.method ?? "GET").toUpperCase() === "DELETE" ||
              (input instanceof Request && input.method === "DELETE"))
          );
        },
        status: 204,
        body: undefined,
      },
    ]);
    restore = stub.restore;

    await expect(unregisterWorkspace("ws-1")).resolves.toBeUndefined();
    expect(stub.calls[0]?.method).toBe("DELETE");
  });

  it("Should throw with server message on unregister failure", async () => {
    const stub = installFetchStub([
      {
        matcher: (input, init) => {
          const url = typeof input === "string" ? input : (input as Request).url;
          return (
            url.includes("/api/workspaces/ws-missing") &&
            ((init?.method ?? "GET").toUpperCase() === "DELETE" ||
              (input instanceof Request && input.method === "DELETE"))
          );
        },
        status: 404,
        body: { code: "not_found", message: "workspace not found" },
      },
    ]);
    restore = stub.restore;

    await expect(unregisterWorkspace("ws-missing")).rejects.toThrow(/workspace not found/);
  });

  // The daemon enforces a CSRF double-submit token on every mutation. These
  // raw-fetch adapters bypass the typed client's CSRF middleware, so they must
  // echo the rc_csrf cookie in the X-rc-CSRF-Token header or the daemon rejects
  // the request with "csrf token is required".
  it("Should send the CSRF token header when registering with a cookie present", async () => {
    setCsrfCookie("csrf-abc");
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/workspaces", "POST"),
        status: 200,
        body: { workspace: { ...workspaceFixture, name: "New WS" } },
      },
    ]);
    restore = stub.restore;

    await registerWorkspace({ name: "New WS", rootDir: "/tmp/new-ws" });
    expect(stub.calls[0]?.headers[CSRF_HEADER_NAME.toLowerCase()]).toBe("csrf-abc");
  });

  it("Should send the CSRF token header when unregistering with a cookie present", async () => {
    setCsrfCookie("csrf-xyz");
    const stub = installFetchStub([
      {
        matcher: (input, init) => {
          const url = typeof input === "string" ? input : (input as Request).url;
          return (
            url.includes("/api/workspaces/ws-1") &&
            ((init?.method ?? "GET").toUpperCase() === "DELETE" ||
              (input instanceof Request && input.method === "DELETE"))
          );
        },
        status: 204,
        body: undefined,
      },
    ]);
    restore = stub.restore;

    await unregisterWorkspace("ws-1");
    expect(stub.calls[0]?.headers[CSRF_HEADER_NAME.toLowerCase()]).toBe("csrf-xyz");
  });

  it("Should omit the CSRF header when no cookie is present", async () => {
    clearCsrfCookie();
    const stub = installFetchStub([
      {
        matcher: matchPath("/api/workspaces", "POST"),
        status: 200,
        body: { workspace: workspaceFixture },
      },
    ]);
    restore = stub.restore;

    await registerWorkspace({ name: "x", rootDir: "/tmp/x" });
    expect(stub.calls[0]?.headers[CSRF_HEADER_NAME.toLowerCase()]).toBeUndefined();
  });
});
