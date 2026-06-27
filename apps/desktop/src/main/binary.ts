import * as fs from "node:fs";
import * as path from "node:path";

/**
 * Resolves the rc binary path using the documented precedence:
 *   1. RC_BINARY environment variable
 *   2. <process.resourcesPath>/bin/rc (bundled inside .app)
 *   3. rc on PATH (via which)
 *
 * Returns the first path that exists and is executable, or throws if none
 * is found so the caller can surface a clear error to the user.
 */
export interface BinaryResolver {
  resolve(): string;
}

export class rcBinaryResolver implements BinaryResolver {
  private readonly resourcesPath: string;
  private readonly env: NodeJS.ProcessEnv;
  private readonly existsFn: (p: string) => boolean;

  constructor(opts?: {
    resourcesPath?: string;
    env?: NodeJS.ProcessEnv;
    existsFn?: (p: string) => boolean;
  }) {
    this.resourcesPath = opts?.resourcesPath ?? process.resourcesPath;
    this.env = opts?.env ?? process.env;
    this.existsFn = opts?.existsFn ?? isExecutable;
  }

  resolve(): string {
    const fromEnv = this.env["RC_BINARY"];
    if (fromEnv && this.existsFn(fromEnv)) {
      return fromEnv;
    }

    const bundled = path.join(this.resourcesPath, "bin", "rc");
    if (this.existsFn(bundled)) {
      return bundled;
    }

    const onPath = resolveFromPath(this.env["PATH"] ?? "");
    if (onPath !== null) {
      return onPath;
    }

    throw new Error(
      "rc binary not found. Set RC_BINARY, bundle the binary in <app>/Contents/Resources/bin/rc, or add rc to PATH."
    );
  }
}

function isExecutable(filePath: string): boolean {
  try {
    fs.accessSync(filePath, fs.constants.X_OK);
    return true;
  } catch {
    return false;
  }
}

function resolveFromPath(pathEnv: string): string | null {
  for (const dir of pathEnv.split(path.delimiter)) {
    if (!dir) continue;
    const candidate = path.join(dir, "rc");
    try {
      fs.accessSync(candidate, fs.constants.X_OK);
      return candidate;
    } catch {
      // not found or not executable in this dir
    }
  }
  return null;
}
