import type { ChildProcess } from "node:child_process";
import { EventEmitter } from "node:events";
import { describe, expect, it, vi } from "vitest";

import { rcDaemonSupervisor } from "./daemon";
import type { BinaryResolver } from "./binary";
import type { DaemonInfoReader, DaemonInfo } from "./info";

const BINARY_PATH = "/usr/local/bin/rc";

function makeBinaryResolver(path: string = BINARY_PATH): BinaryResolver {
  return { resolve: () => path };
}

function makeInfoReader(info: DaemonInfo | null = null): DaemonInfoReader {
  return {
    read: () => info,
    infoPath: () => "/home/.rc/daemon/daemon.json",
  };
}

function makeDaemonInfo(port = 2323): DaemonInfo {
  return {
    pid: 999,
    http_port: port,
    socket_path: "/tmp/rc.sock",
    state: "healthy",
    started_at: "2026-01-01T00:00:00Z",
  };
}

function makeFakeProcess(): ChildProcess {
  const em = new EventEmitter() as ChildProcess;
  (em as unknown as { kill: () => boolean }).kill = () => true;
  return em;
}

function makeSpawnFn(proc: ChildProcess) {
  return vi.fn(() => proc);
}

function healthyFetch(_port = 2323): typeof globalThis.fetch {
  return vi.fn(async (input: RequestInfo | URL) => {
    const url = typeof input === "string" ? input : input.toString();
    if (url.includes("/api/daemon/health")) {
      return new Response(null, { status: 200 });
    }
    if (url.includes("/api/daemon/stop")) {
      return new Response(null, { status: 200 });
    }
    return new Response(null, { status: 404 });
  }) as unknown as typeof globalThis.fetch;
}

function unhealthyFetch(): typeof globalThis.fetch {
  return vi.fn(async () => {
    throw new Error("connection refused");
  }) as unknown as typeof globalThis.fetch;
}

describe("rcDaemonSupervisor — attach vs spawn decision", () => {
  it("Should attach to a running daemon without spawning when health probe succeeds", async () => {
    const info = makeDaemonInfo();
    const proc = makeFakeProcess();
    const spawnFn = makeSpawnFn(proc);

    const supervisor = new rcDaemonSupervisor({
      binaryResolver: makeBinaryResolver(),
      infoReader: makeInfoReader(info),
      spawnFn: spawnFn as unknown as typeof import("node:child_process").spawn,
      fetchFn: healthyFetch(),
      pollIntervalMs: 10,
      startupTimeoutMs: 500,
    });

    await supervisor.start();

    expect(supervisor.state).toBe("healthy");
    expect(spawnFn).not.toHaveBeenCalled();
    expect(supervisor.httpPort).toBe(2323);
  });

  it("Should spawn rc daemon start when health probe fails (no running daemon)", async () => {
    const proc = makeFakeProcess();
    const spawnFn = makeSpawnFn(proc);

    // infoReader returns null initially (daemon not running), then returns info after spawn.
    let spawnCalled = false;
    const originalSpawn = spawnFn.getMockImplementation();
    spawnFn.mockImplementation((...args) => {
      spawnCalled = true;
      return originalSpawn!(...args);
    });

    const infoReader: DaemonInfoReader = {
      read: () => (spawnCalled ? makeDaemonInfo() : null),
      infoPath: () => "/home/.rc/daemon/daemon.json",
    };

    const fetchFn = vi.fn(async (input: RequestInfo | URL) => {
      const url = typeof input === "string" ? input : input.toString();
      if (url.includes("/api/daemon/health")) {
        return spawnCalled
          ? new Response(null, { status: 200 })
          : (() => {
              throw new Error("connection refused");
            })();
      }
      return new Response(null, { status: 200 });
    }) as unknown as typeof globalThis.fetch;

    const supervisor = new rcDaemonSupervisor({
      binaryResolver: makeBinaryResolver(),
      infoReader,
      spawnFn: spawnFn as unknown as typeof import("node:child_process").spawn,
      fetchFn,
      pollIntervalMs: 10,
      startupTimeoutMs: 500,
    });

    await supervisor.start();

    expect(supervisor.state).toBe("healthy");
    expect(spawnFn).toHaveBeenCalledWith(BINARY_PATH, ["daemon", "start"], expect.any(Object));
  });
});

describe("rcDaemonSupervisor — graceful stop policy", () => {
  it("Should POST /api/daemon/stop on quit only when it owns the daemon", async () => {
    const proc = makeFakeProcess();
    const spawnFn = makeSpawnFn(proc);

    let spawnCalled = false;
    const originalSpawn = spawnFn.getMockImplementation();
    spawnFn.mockImplementation((...args) => {
      spawnCalled = true;
      return originalSpawn!(...args);
    });

    const infoReader: DaemonInfoReader = {
      read: () => (spawnCalled ? makeDaemonInfo() : null),
      infoPath: () => "/home/.rc/daemon/daemon.json",
    };

    const fetchSpy = vi.fn(async (input: RequestInfo | URL) => {
      const url = typeof input === "string" ? input : input.toString();
      if (url.includes("/api/daemon/health")) {
        return spawnCalled
          ? new Response(null, { status: 200 })
          : (() => {
              throw new Error("not running");
            })();
      }
      return new Response(null, { status: 200 });
    }) as unknown as typeof globalThis.fetch;

    const supervisor = new rcDaemonSupervisor({
      binaryResolver: makeBinaryResolver(),
      infoReader,
      spawnFn: spawnFn as unknown as typeof import("node:child_process").spawn,
      fetchFn: fetchSpy,
      pollIntervalMs: 10,
      startupTimeoutMs: 500,
    });

    await supervisor.start();
    expect(supervisor.state).toBe("healthy");

    await supervisor.stop();

    const stopCalls = (fetchSpy as ReturnType<typeof vi.fn>).mock.calls.filter(
      (args: unknown[]) => {
        const input = args[0] as RequestInfo | URL;
        const url = typeof input === "string" ? input : input.toString();
        return url.includes("/api/daemon/stop");
      }
    );
    expect(stopCalls.length).toBeGreaterThanOrEqual(1);
    expect(supervisor.state).toBe("stopped");
  });

  it("Should NOT POST /api/daemon/stop when it only attached (does not own daemon)", async () => {
    const fetchSpy = vi.fn(healthyFetch());

    const supervisor = new rcDaemonSupervisor({
      binaryResolver: makeBinaryResolver(),
      infoReader: makeInfoReader(makeDaemonInfo()),
      spawnFn: vi.fn() as unknown as typeof import("node:child_process").spawn,
      fetchFn: fetchSpy as unknown as typeof globalThis.fetch,
      pollIntervalMs: 10,
      startupTimeoutMs: 500,
    });

    await supervisor.start();
    expect(supervisor.state).toBe("healthy");

    await supervisor.stop();

    const stopCalls = (fetchSpy as ReturnType<typeof vi.fn>).mock.calls.filter(
      (args: unknown[]) => {
        const input = args[0] as RequestInfo | URL;
        const url = typeof input === "string" ? input : input.toString();
        return url.includes("/api/daemon/stop");
      }
    );
    expect(stopCalls).toHaveLength(0);
    expect(supervisor.state).toBe("stopped");
  });
});

describe("rcDaemonSupervisor — bounded restart backoff", () => {
  it("Should transition to unhealthy after exhausting all restart attempts", async () => {
    const spawnFn = vi.fn(() => makeFakeProcess());

    const supervisor = new rcDaemonSupervisor({
      binaryResolver: makeBinaryResolver(),
      infoReader: makeInfoReader(null),
      spawnFn: spawnFn as unknown as typeof import("node:child_process").spawn,
      fetchFn: unhealthyFetch(),
      pollIntervalMs: 1,
      startupTimeoutMs: 20,
      restartBackoffMs: [1, 1],
    });

    await supervisor.start();

    expect(supervisor.state).toBe("unhealthy");
    expect(spawnFn.mock.calls.length).toBeGreaterThanOrEqual(2);
  });
});
