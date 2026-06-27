import { QueryClientProvider } from "@tanstack/react-query";
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from "@tanstack/react-router";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";
import { afterEach, describe, expect, it } from "vitest";

import { createTestQueryClient, installFetchStub, matchPath } from "@/test/utils";
import type { Workspace } from "@/systems/app-shell";

import { ExecView } from "./exec-view";

const workspace = {
  id: "ws-1",
  name: "one",
  root_dir: "/tmp/ws",
  read_only: false,
} as Workspace;

const startedRun = {
  run: {
    run_id: "run-exec",
    mode: "exec",
    presentation_mode: "detach",
    workspace_id: "ws-1",
    started_at: "2026-01-01T00:00:00Z",
    status: "queued",
  },
};

async function renderExecView() {
  const rootRoute = createRootRoute();
  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/",
    component: function IndexRouteComponent(): ReactElement {
      return <ExecView activeWorkspace={workspace} />;
    },
  });
  const runDetailRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/runs/$runId",
    component: function RunDetailStub(): ReactElement {
      return <div data-testid="run-detail-stub" />;
    },
  });
  const router = createRouter({
    routeTree: rootRoute.addChildren([indexRoute, runDetailRoute]),
    history: createMemoryHistory({ initialEntries: ["/"] }),
    defaultPreload: false,
  });
  await router.load();
  await act(async () => {
    render(
      <QueryClientProvider client={createTestQueryClient()}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    );
    await Promise.resolve();
  });
}

describe("ExecView interactive flag", () => {
  let restore: (() => void) | null = null;

  afterEach(() => {
    restore?.();
    restore = null;
  });

  it("Should POST interactive: true when the Interactive toggle is checked", async () => {
    const stub = installFetchStub([
      { matcher: matchPath("/api/catalog/agents"), status: 200, body: { agents: [] } },
      { matcher: matchPath("/api/exec", "POST"), status: 201, body: startedRun },
    ]);
    restore = stub.restore;

    await renderExecView();
    await userEvent.type(screen.getByTestId("exec-prompt"), "/brainstorm");
    await userEvent.click(screen.getByTestId("exec-interactive"));
    await userEvent.click(screen.getByTestId("exec-submit"));

    const execCall = stub.calls.find(
      call => call.method === "POST" && call.url.includes("/api/exec")
    );
    expect(execCall, "expected a POST to /api/exec").toBeTruthy();
    const body = JSON.parse(execCall?.body ?? "{}");
    expect(body.interactive).toBe(true);
  });

  it("Should omit interactive when the toggle is left unchecked", async () => {
    const stub = installFetchStub([
      { matcher: matchPath("/api/catalog/agents"), status: 200, body: { agents: [] } },
      { matcher: matchPath("/api/exec", "POST"), status: 201, body: startedRun },
    ]);
    restore = stub.restore;

    await renderExecView();
    await userEvent.type(screen.getByTestId("exec-prompt"), "/brainstorm");
    await userEvent.click(screen.getByTestId("exec-submit"));

    const execCall = stub.calls.find(
      call => call.method === "POST" && call.url.includes("/api/exec")
    );
    const body = JSON.parse(execCall?.body ?? "{}");
    expect(body.interactive).toBeUndefined();
  });
});
