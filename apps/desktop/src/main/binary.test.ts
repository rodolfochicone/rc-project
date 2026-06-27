import * as fs from "node:fs";
import * as os from "node:os";
import * as path from "node:path";
import { afterEach, beforeEach, describe, expect, it } from "vitest";

import { rcBinaryResolver } from "./binary";

function makeTempDir(): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), "rc-binary-test-"));
}

function writeExecutable(dir: string, name: string): string {
  const p = path.join(dir, name);
  fs.writeFileSync(p, "#!/bin/sh\n", { mode: 0o755 });
  return p;
}

describe("rcBinaryResolver", () => {
  let tmpDir: string;

  beforeEach(() => {
    tmpDir = makeTempDir();
  });

  afterEach(() => {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  });

  it("Should prefer RC_BINARY env var when the path is executable", () => {
    const binary = writeExecutable(tmpDir, "rc");
    const resolver = new rcBinaryResolver({
      env: { RC_BINARY: binary },
      resourcesPath: tmpDir,
    });
    expect(resolver.resolve()).toBe(binary);
  });

  it("Should skip RC_BINARY when the file is not executable and fall through", () => {
    const notExec = path.join(tmpDir, "rc-not-exec");
    fs.writeFileSync(notExec, "");
    const bundledDir = path.join(tmpDir, "resources", "bin");
    fs.mkdirSync(bundledDir, { recursive: true });
    const bundled = writeExecutable(bundledDir, "rc");

    const resolver = new rcBinaryResolver({
      env: { RC_BINARY: notExec },
      resourcesPath: path.join(tmpDir, "resources"),
    });
    expect(resolver.resolve()).toBe(bundled);
  });

  it("Should resolve bundled binary under resourcesPath/bin/rc when env is absent", () => {
    const bundledDir = path.join(tmpDir, "bin");
    fs.mkdirSync(bundledDir);
    const bundled = writeExecutable(bundledDir, "rc");

    const resolver = new rcBinaryResolver({
      env: {},
      resourcesPath: tmpDir,
    });
    expect(resolver.resolve()).toBe(bundled);
  });

  it("Should resolve from PATH when env and bundled are absent", () => {
    const pathDir = path.join(tmpDir, "pathdir");
    fs.mkdirSync(pathDir);
    const onPath = writeExecutable(pathDir, "rc");

    const resolver = new rcBinaryResolver({
      env: { PATH: pathDir },
      resourcesPath: path.join(tmpDir, "no-resources"),
    });
    expect(resolver.resolve()).toBe(onPath);
  });

  it("Should throw when no rc binary is found in any location", () => {
    const resolver = new rcBinaryResolver({
      env: {},
      resourcesPath: path.join(tmpDir, "no-resources"),
    });
    expect(() => resolver.resolve()).toThrow(/rc binary not found/);
  });

  it("Should respect resolution order: RC_BINARY before bundled before PATH", () => {
    const envBinary = writeExecutable(tmpDir, "rc-env");
    const bundledDir = path.join(tmpDir, "bin");
    fs.mkdirSync(bundledDir);
    writeExecutable(bundledDir, "rc");
    const pathDir = path.join(tmpDir, "pathdir");
    fs.mkdirSync(pathDir);
    writeExecutable(pathDir, "rc");

    const resolver = new rcBinaryResolver({
      env: { RC_BINARY: envBinary, PATH: pathDir },
      resourcesPath: tmpDir,
    });
    expect(resolver.resolve()).toBe(envBinary);
  });
});
