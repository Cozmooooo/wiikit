// Wails runtime stubs for TypeScript — real bindings injected at runtime by Wails
declare global {
  interface Window {
    go: {
      main: {
        App: {
          ListDrives(): Promise<import('./stores/state').DriveInfo[]>
          ScanLibrary(roots: string[]): Promise<import('./stores/state').Game[]>
          OpenDirectoryDialog(): Promise<string>
          PreviewFormat(d: import('./stores/state').DriveInfo, label: string): Promise<string>
          StartInstall(
            games: import('./stores/state').Game[],
            usb: import('./stores/state').DriveInfo,
            sd: import('./stores/state').DriveInfo
          ): Promise<void>
        }
      }
    }
    runtime: {
      EventsOn(event: string, cb: (...args: unknown[]) => void): void
      EventsOff(...events: string[]): void
    }
  }
}
export {}
