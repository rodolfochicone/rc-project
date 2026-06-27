import * as fs from "node:fs";
import * as os from "node:os";
import * as path from "node:path";

/** Shape of ~/.rc/daemon/daemon.json written by the Go daemon. */
export interface DaemonInfo {
  pid: number;
  http_port: number;
  socket_path: string;
  state: string;
  started_at: string;
  version?: string;
}

export interface DaemonInfoReader {
  read(): DaemonInfo | null;
  infoPath(): string;
}

export class rcDaemonInfoReader implements DaemonInfoReader {
  private readonly homeDirFn: () => string;

  constructor(opts?: { homeDirFn?: () => string }) {
    this.homeDirFn = opts?.homeDirFn ?? os.homedir;
  }

  infoPath(): string {
    return path.join(this.homeDirFn(), ".rc", "daemon", "daemon.json");
  }

  read(): DaemonInfo | null {
    const p = this.infoPath();
    let raw: string;
    try {
      raw = fs.readFileSync(p, "utf-8");
    } catch {
      return null;
    }
    const parsed = JSON.parse(raw) as Partial<DaemonInfo>;
    if (
      typeof parsed.pid !== "number" ||
      typeof parsed.http_port !== "number" ||
      typeof parsed.socket_path !== "string" ||
      typeof parsed.state !== "string" ||
      typeof parsed.started_at !== "string"
    ) {
      throw new Error(`daemon.json at ${p} is missing required fields`);
    }
    return parsed as DaemonInfo;
  }
}
