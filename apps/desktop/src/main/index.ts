import * as path from "node:path";

import { app, BrowserWindow, dialog, ipcMain, Menu, shell } from "electron";

import { rcBinaryResolver } from "./binary";
import { rcDaemonSupervisor } from "./daemon";
import { rcDaemonInfoReader } from "./info";
import { buildTrayMenu, createTray } from "./tray";

const supervisor = new rcDaemonSupervisor({
  binaryResolver: new rcBinaryResolver(),
  infoReader: new rcDaemonInfoReader(),
});

let mainWindow: BrowserWindow | null = null;

function createWindow(): void {
  const win = new BrowserWindow({
    width: 1280,
    height: 800,
    webPreferences: {
      preload: path.join(__dirname, "..", "preload", "index.js"),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  });

  mainWindow = win;

  const port = supervisor.httpPort;
  if (port !== null) {
    void win.loadURL(`http://127.0.0.1:${port}`);
  } else {
    void win.loadURL("about:blank");
  }

  win.on("closed", () => {
    mainWindow = null;
  });
}

// Native directory picker for the renderer (browsers cannot resolve an absolute
// filesystem path). Returns the chosen directory, or null if the user cancels.
function registerDirectoryPicker(): void {
  ipcMain.handle("rc:select-directory", async event => {
    const win = BrowserWindow.fromWebContents(event.sender);
    const options: Electron.OpenDialogOptions = {
      title: "Select workspace folder",
      properties: ["openDirectory", "createDirectory"],
    };
    const result = win
      ? await dialog.showOpenDialog(win, options)
      : await dialog.showOpenDialog(options);
    if (result.canceled) {
      return null;
    }
    return result.filePaths[0] ?? null;
  });
}

function buildAppMenu(): void {
  const template: Electron.MenuItemConstructorOptions[] = [
    {
      label: "rc",
      submenu: [
        { role: "about" },
        { type: "separator" },
        { role: "quit", accelerator: "CmdOrCtrl+Q" },
      ],
    },
    {
      label: "View",
      submenu: [
        {
          label: "Reload",
          accelerator: "CmdOrCtrl+R",
          click: () => {
            mainWindow?.webContents.reload();
          },
        },
      ],
    },
    {
      label: "Window",
      submenu: [{ role: "minimize" }, { role: "zoom" }, { type: "separator" }, { role: "front" }],
    },
    {
      role: "help",
      submenu: [
        {
          label: "rc Documentation",
          click: async () => {
            await shell.openExternal("https://github.com/rodolfochicone/rc-project");
          },
        },
      ],
    },
  ];
  Menu.setApplicationMenu(Menu.buildFromTemplate(template));
}

const gotLock = app.requestSingleInstanceLock();
if (!gotLock) {
  app.quit();
} else {
  app.on("second-instance", () => {
    if (mainWindow) {
      if (mainWindow.isMinimized()) mainWindow.restore();
      mainWindow.focus();
    }
  });

  app.on("ready", () => {
    buildAppMenu();
    registerDirectoryPicker();

    const tray = createTray(
      supervisor.state,
      () => {
        if (!mainWindow) {
          createWindow();
        } else {
          mainWindow.show();
        }
      },
      () => {
        void supervisor.stop().then(() => supervisor.start());
      }
    );

    supervisor.on("stateChange", state => {
      tray.setContextMenu(
        buildTrayMenu(
          state,
          () => {
            if (!mainWindow) {
              createWindow();
            } else {
              mainWindow.show();
            }
          },
          () => {
            void supervisor.stop().then(() => supervisor.start());
          },
          () => {
            app.quit();
          }
        )
      );
    });

    void supervisor.start().then(() => {
      createWindow();
    });
  });

  app.on("activate", () => {
    if (mainWindow === null) {
      createWindow();
    }
  });

  let quitting = false;
  app.on("before-quit", event => {
    if (quitting) {
      return;
    }
    // Defer the actual quit until the owned daemon has been stopped gracefully;
    // supervisor.stop() is async (POST /api/daemon/stop) and would otherwise be
    // cut off when the process exits, orphaning the daemon.
    event.preventDefault();
    quitting = true;
    void supervisor.stop().finally(() => {
      app.exit(0);
    });
  });

  app.on("window-all-closed", () => {
    // Keep the app running in the tray on macOS; quit on other platforms.
    if (process.platform !== "darwin") {
      app.quit();
    }
  });
}
