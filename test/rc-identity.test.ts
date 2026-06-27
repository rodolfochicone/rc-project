import { readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

const rootDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");

async function readText(relativePath: string): Promise<string> {
  return readFile(path.join(rootDir, relativePath), "utf8");
}

async function readJSON<T>(relativePath: string): Promise<T> {
  return JSON.parse(await readText(relativePath)) as T;
}

interface PackageJSON {
  name?: string;
  dependencies?: Record<string, string>;
}

// T16, F2.6: web package names must use the rc scope, not compozy.
describe("rc web package identity (T16, F2.6)", () => {
  it("root package.json is named rc-project, not compozy-harness", async () => {
    const pkg = await readJSON<PackageJSON>("package.json");
    expect(pkg.name).not.toBe("compozy-harness");
    expect(pkg.name).toBe("rc-project");
  });

  it("web/package.json is named rc-web, not compozy-web", async () => {
    const pkg = await readJSON<PackageJSON>("web/package.json");
    expect(pkg.name).not.toBe("compozy-web");
    expect(pkg.name).toBe("rc-web");
  });

  it("packages/ui is scoped @rodolfochicone/ui, not @compozy-tech/ui", async () => {
    const pkg = await readJSON<PackageJSON>("packages/ui/package.json");
    expect(pkg.name).not.toBe("@compozy-tech/ui");
    expect(pkg.name).toBe("@rodolfochicone/ui");
  });

  it("web/package.json depends on @rodolfochicone/ui workspace, not @compozy-tech/ui", async () => {
    const pkg = await readJSON<PackageJSON>("web/package.json");
    expect(pkg.dependencies?.["@compozy-tech/ui"]).toBeUndefined();
    expect(pkg.dependencies?.["@rodolfochicone/ui"]).toBe("workspace:*");
  });

  it("web/components.json shadcn aliases use @rodolfochicone/ui, not @compozy-tech/ui", async () => {
    const componentsJson = await readText("web/components.json");
    expect(componentsJson).not.toContain("@compozy-tech/ui");
    expect(componentsJson).toContain("@rodolfochicone/ui");
  });
});

// T19, F3.2, AC6: web brand color tokens must be orange/amber, not rc green.
describe("rc web brand palette (T19, F3.2, AC6)", () => {
  it("packages/ui/src/tokens.css uses orange brand token #f26b21, not green #d6f24a", async () => {
    const css = await readText("packages/ui/src/tokens.css");
    expect(css).not.toContain("#d6f24a");
    expect(css).not.toContain("#D6F24A");
    expect(css.toLowerCase()).toContain("--brand: #f26b21");
  });

  it("web/src/styles.css uses orange rgba gradient, not green rgba(214, 242, 74", async () => {
    const css = await readText("web/src/styles.css");
    expect(css).not.toContain("rgba(214, 242, 74");
    expect(css).not.toContain("rgba(214,242,74");
    expect(css).toContain("rgba(242, 107, 33");
  });
});

// T17, AC5 (partial): TS/TSX files must contain no compozy identifier strings.
describe("rc TS/TSX residual-compozy scan (T17, AC5 partial)", () => {
  it("web/src/generated/ openapi type file is named rc-openapi.d.ts, not compozy-openapi.d.ts", async () => {
    const { existsSync } = await import("node:fs");
    const compozyPath = path.join(rootDir, "web/src/generated/compozy-openapi.d.ts");
    const rcPath = path.join(rootDir, "web/src/generated/rc-openapi.d.ts");

    expect(
      existsSync(compozyPath),
      "web/src/generated/compozy-openapi.d.ts still exists — not renamed to rc-openapi.d.ts"
    ).toBe(false);
    expect(
      existsSync(rcPath),
      "web/src/generated/rc-openapi.d.ts does not exist — codegen target not renamed"
    ).toBe(true);
  });

  it("scripts/codegen.mjs references rc-daemon.json input and rc-openapi.d.ts output", async () => {
    const codegen = await readText("scripts/codegen.mjs");
    expect(codegen).not.toContain("compozy-daemon.json");
    expect(codegen).not.toContain("compozy-openapi.d.ts");
    expect(codegen).toContain("rc-daemon.json");
    expect(codegen).toContain("rc-openapi.d.ts");
  });
});
