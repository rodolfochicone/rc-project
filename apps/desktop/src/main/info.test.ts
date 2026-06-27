import * as fs from "node:fs";
import * as os from "node:os";
import * as path from "node:path";
import { afterEach, beforeEach, describe, expect, it } from "vitest";

import { rcDaemonInfoReader, type DaemonInfo } from "./info";

function makeTempDir(): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), "rc-info-test-"));
}

const validInfo: DaemonInfo = {
  pid: 12345,
  http_port: 2323,
  socket_path: "/tmp/rc/daemon.sock",
  state: "healthy",
  started_at: "2026-01-01T00:00:00Z",
  version: "0.2.4",
};

describe("rcDaemonInfoReader", () => {
  let tmpDir: string;

  beforeEach(() => {
    tmpDir = makeTempDir();
  });

  afterEach(() => {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  });

  it("Should return null when daemon.json does not exist", () => {
    const reader = new rcDaemonInfoReader({ homeDirFn: () => tmpDir });
    expect(reader.read()).toBeNull();
  });

  it("Should parse a valid daemon.json and return DaemonInfo", () => {
    const daemonDir = path.join(tmpDir, ".rc", "daemon");
    fs.mkdirSync(daemonDir, { recursive: true });
    fs.writeFileSync(path.join(daemonDir, "daemon.json"), JSON.stringify(validInfo));

    const reader = new rcDaemonInfoReader({ homeDirFn: () => tmpDir });
    const info = reader.read();
    expect(info).not.toBeNull();
    expect(info?.pid).toBe(12345);
    expect(info?.http_port).toBe(2323);
    expect(info?.socket_path).toBe("/tmp/rc/daemon.sock");
    expect(info?.state).toBe("healthy");
    expect(info?.started_at).toBe("2026-01-01T00:00:00Z");
    expect(info?.version).toBe("0.2.4");
  });

  it("Should throw when daemon.json is missing required fields", () => {
    const daemonDir = path.join(tmpDir, ".rc", "daemon");
    fs.mkdirSync(daemonDir, { recursive: true });
    fs.writeFileSync(path.join(daemonDir, "daemon.json"), JSON.stringify({ pid: 99 }));

    const reader = new rcDaemonInfoReader({ homeDirFn: () => tmpDir });
    expect(() => reader.read()).toThrow(/missing required fields/);
  });

  it("Should return the correct infoPath based on home dir", () => {
    const reader = new rcDaemonInfoReader({ homeDirFn: () => "/custom/home" });
    expect(reader.infoPath()).toBe("/custom/home/.rc/daemon/daemon.json");
  });

  it("Should work without version field (optional)", () => {
    const daemonDir = path.join(tmpDir, ".rc", "daemon");
    fs.mkdirSync(daemonDir, { recursive: true });
    const withoutVersion = { ...validInfo };
    delete (withoutVersion as Partial<DaemonInfo>).version;
    fs.writeFileSync(path.join(daemonDir, "daemon.json"), JSON.stringify(withoutVersion));

    const reader = new rcDaemonInfoReader({ homeDirFn: () => tmpDir });
    const info = reader.read();
    expect(info).not.toBeNull();
    expect(info?.version).toBeUndefined();
  });
});
