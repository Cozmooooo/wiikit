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
          SearchDeals(query: string, maxPrice: number, minPrice: number): Promise<{
            listings: Array<{
              id: string; title: string; price: number; currency: string;
              condition: string; url: string; imageUrl: string; seller: string;
              location: string; source: string; score: number;
            }>;
            ebayError?: string;
            facebookError?: string;
          }>
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
