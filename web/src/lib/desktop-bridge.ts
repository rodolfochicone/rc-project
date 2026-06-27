/**
 * Bridge exposed by the Electron preload (`window.rcdesktop`). It is absent when
 * the UI runs in a plain browser, so every consumer must treat it as optional
 * and fall back to typed input.
 */
export interface rcDesktopBridge {
  version?: string;
  selectDirectory?: () => Promise<string | null>;
}

declare global {
  interface Window {
    rcdesktop?: rcDesktopBridge;
  }
}

export function getDesktopBridge(): rcDesktopBridge | null {
  if (typeof window === "undefined") {
    return null;
  }
  return window.rcdesktop ?? null;
}

/** True when the host can open a native directory picker (Electron only). */
export function canPickDirectory(): boolean {
  return typeof getDesktopBridge()?.selectDirectory === "function";
}
