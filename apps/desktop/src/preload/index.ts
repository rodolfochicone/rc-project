import { contextBridge, ipcRenderer } from "electron";

/**
 * Minimal preload: exposes the desktop version and a native directory picker to
 * the renderer. No broad Node.js APIs are exposed; contextIsolation + sandbox
 * keep the renderer safe, and only the explicit IPC channel is bridged.
 */
contextBridge.exposeInMainWorld("rcdesktop", {
  version: process.env["RC_DESKTOP_VERSION"] ?? "dev",
  selectDirectory: (): Promise<string | null> => ipcRenderer.invoke("rc:select-directory"),
} as const);
