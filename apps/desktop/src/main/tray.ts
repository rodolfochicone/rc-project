import { app, Menu, nativeImage, Tray } from "electron";

import type { DaemonState } from "./daemon";

// Escale brand mark at 32x32, treated as @2x so it renders as a crisp 16pt
// menu-bar icon on Retina displays. Embedded as base64 so it ships with the
// bundle without resolving a file path that differs between dev and packaged
// builds. Daemon state is surfaced through the menu label, not the icon color.
const TRAY_ICON_BASE64 =
  "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";

function trayIcon(): Electron.NativeImage {
  return nativeImage.createFromBuffer(Buffer.from(TRAY_ICON_BASE64, "base64"), {
    scaleFactor: 2,
  });
}

function stateLabel(state: DaemonState): string {
  const labels: Record<DaemonState, string> = {
    starting: "Starting…",
    healthy: "Running",
    unhealthy: "Unhealthy",
    stopped: "Stopped",
  };
  return labels[state];
}

export function buildTrayMenu(
  state: DaemonState,
  onShow: () => void,
  onRestart: () => void,
  onQuit: () => void
): Electron.Menu {
  return Menu.buildFromTemplate([
    { label: `rc — ${stateLabel(state)}`, enabled: false },
    { type: "separator" },
    { label: "Show", click: onShow },
    { label: "Restart daemon", click: onRestart },
    { type: "separator" },
    { label: "Quit", click: onQuit },
  ]);
}

export function createTray(state: DaemonState, onShow: () => void, onRestart: () => void): Tray {
  const tray = new Tray(trayIcon());
  tray.setToolTip("rc");
  tray.setContextMenu(
    buildTrayMenu(state, onShow, onRestart, () => {
      app.quit();
    })
  );
  return tray;
}
