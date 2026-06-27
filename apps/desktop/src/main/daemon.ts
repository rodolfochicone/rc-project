import * as child_process from "node:child_process";
import * as events from "node:events";

import type { BinaryResolver } from "./binary";
import type { DaemonInfoReader } from "./info";

export type DaemonState = "starting" | "healthy" | "unhealthy" | "stopped";

export interface DaemonSupervisor extends events.EventEmitter {
  readonly state: DaemonState;
  readonly httpPort: number | null;
  start(): Promise<void>;
  stop(): Promise<void>;
}

export interface SupervisorOptions {
  binaryResolver: BinaryResolver;
  infoReader: DaemonInfoReader;
  /** Override child_process.spawn for testing. */
  spawnFn?: typeof child_process.spawn;
  /** Override globalThis.fetch for testing. */
  fetchFn?: typeof globalThis.fetch;
  /** Startup timeout in ms (default 30 000). */
  startupTimeoutMs?: number;
  /** Poll interval in ms for health checks (default 250). */
  pollIntervalMs?: number;
  /** Backoff intervals in ms for crash restarts (default 1s/2s/4s/8s/16s). */
  restartBackoffMs?: number[];
}

const DEFAULT_BACKOFF_MS = [1000, 2000, 4000, 8000, 16000];
const DEFAULT_STARTUP_TIMEOUT_MS = 30_000;
const DEFAULT_POLL_INTERVAL_MS = 250;

/**
 * Supervisor owns the daemon lifecycle:
 * - attach-or-spawn on start()
 * - 3s health polling while running
 * - bounded exponential-backoff crash restart (max 5 attempts)
 * - graceful POST /api/daemon/stop on quit only when it spawned the daemon
 */
export class rcDaemonSupervisor extends events.EventEmitter implements DaemonSupervisor {
  private _state: DaemonState = "starting";
  private _httpPort: number | null = null;
  private _ownsDaemon = false;
  private _stopped = false;
  private _daemonProcess: child_process.ChildProcess | null = null;

  private readonly binaryResolver: BinaryResolver;
  private readonly infoReader: DaemonInfoReader;
  private readonly spawnFn: typeof child_process.spawn;
  private readonly fetchFn: typeof globalThis.fetch;
  private readonly startupTimeoutMs: number;
  private readonly pollIntervalMs: number;
  private readonly restartBackoffMs: number[];

  constructor(opts: SupervisorOptions) {
    super();
    this.binaryResolver = opts.binaryResolver;
    this.infoReader = opts.infoReader;
    this.spawnFn = opts.spawnFn ?? child_process.spawn;
    this.fetchFn = opts.fetchFn ?? globalThis.fetch.bind(globalThis);
    this.startupTimeoutMs = opts.startupTimeoutMs ?? DEFAULT_STARTUP_TIMEOUT_MS;
    this.pollIntervalMs = opts.pollIntervalMs ?? DEFAULT_POLL_INTERVAL_MS;
    this.restartBackoffMs = opts.restartBackoffMs ?? DEFAULT_BACKOFF_MS;
  }

  get state(): DaemonState {
    return this._state;
  }

  get httpPort(): number | null {
    return this._httpPort;
  }

  async start(): Promise<void> {
    this._stopped = false;
    await this._attachOrSpawn();
  }

  async stop(): Promise<void> {
    this._stopped = true;
    if (this._ownsDaemon && this._httpPort !== null) {
      await this._gracefulStop();
    }
    this._setState("stopped");
    if (this._daemonProcess) {
      this._daemonProcess.kill("SIGTERM");
      this._daemonProcess = null;
    }
  }

  private async _attachOrSpawn(): Promise<void> {
    const healthy = await this._probeHealth();
    if (healthy) {
      const info = this.infoReader.read();
      this._httpPort = info?.http_port ?? null;
      this._ownsDaemon = false;
      this._setState("healthy");
      return;
    }
    await this._spawnDaemon(0);
  }

  private async _spawnDaemon(attempt: number): Promise<void> {
    if (this._stopped) return;

    const binary = this.binaryResolver.resolve();
    const proc = this.spawnFn(binary, ["daemon", "start"], {
      detached: false,
      stdio: "ignore",
    });
    this._daemonProcess = proc;
    this._ownsDaemon = true;

    const started = await this._waitForHealthy();
    if (started) {
      const info = this.infoReader.read();
      this._httpPort = info?.http_port ?? null;
      this._setState("healthy");
      proc.once("exit", () => {
        if (!this._stopped) {
          void this._handleCrash(attempt + 1);
        }
      });
      return;
    }

    if (attempt < this.restartBackoffMs.length) {
      const delay = this.restartBackoffMs[attempt] ?? 1000;
      await sleep(delay);
      await this._spawnDaemon(attempt + 1);
    } else {
      this._setState("unhealthy");
    }
  }

  private async _handleCrash(attempt: number): Promise<void> {
    if (this._stopped) return;
    this._setState("unhealthy");
    if (attempt >= this.restartBackoffMs.length) {
      return;
    }
    const delay = this.restartBackoffMs[attempt - 1] ?? 1000;
    await sleep(delay);
    if (!this._stopped) {
      await this._spawnDaemon(attempt);
    }
  }

  private async _probeHealth(): Promise<boolean> {
    if (this._httpPort === null) {
      const info = this.infoReader.read();
      if (!info) return false;
      this._httpPort = info.http_port;
    }
    return this._callHealth(this._httpPort);
  }

  private async _waitForHealthy(): Promise<boolean> {
    const deadline = Date.now() + this.startupTimeoutMs;
    while (Date.now() < deadline) {
      if (this._stopped) return false;
      const info = this.infoReader.read();
      if (info) {
        this._httpPort = info.http_port;
        const ok = await this._callHealth(info.http_port);
        if (ok) return true;
      }
      await sleep(this.pollIntervalMs);
    }
    return false;
  }

  private async _callHealth(port: number): Promise<boolean> {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), 2000);
    try {
      const res = await this.fetchFn(`http://127.0.0.1:${port}/api/daemon/health`, {
        signal: controller.signal,
      });
      return res.ok;
    } catch {
      return false;
    } finally {
      clearTimeout(timer);
    }
  }

  private async _gracefulStop(): Promise<void> {
    if (this._httpPort === null) return;
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), 5000);
    try {
      await this.fetchFn(`http://127.0.0.1:${this._httpPort}/api/daemon/stop`, {
        method: "POST",
        signal: controller.signal,
      });
    } catch {
      // best effort
    } finally {
      clearTimeout(timer);
    }
  }

  private _setState(s: DaemonState): void {
    if (this._state !== s) {
      this._state = s;
      this.emit("stateChange", s);
    }
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}
