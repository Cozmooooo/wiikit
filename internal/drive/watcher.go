package drive

import (
	"context"
	"time"
)

// DriveEvent is emitted by WatchDrives when a removable drive is inserted or removed.
type DriveEvent struct {
	Type  string    `json:"type"`  // "inserted" | "removed"
	Drive DriveInfo `json:"drive"`
}

// WatchDrives polls ListDrives every interval and emits events on the returned
// channel. The channel is closed when ctx is cancelled.
func WatchDrives(ctx context.Context, interval time.Duration) <-chan DriveEvent {
	ch := make(chan DriveEvent, 8)
	go func() {
		defer close(ch)
		var prev []DriveInfo
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			curr, err := ListDrives()
			if err != nil {
				continue
			}
			for _, d := range curr {
				if !hasDrive(prev, d.Path) {
					select {
					case ch <- DriveEvent{Type: "inserted", Drive: d}:
					default:
					}
				}
			}
			for _, d := range prev {
				if !hasDrive(curr, d.Path) {
					select {
					case ch <- DriveEvent{Type: "removed", Drive: d}:
					default:
					}
				}
			}
			prev = curr
		}
	}()
	return ch
}

func hasDrive(drives []DriveInfo, path string) bool {
	for _, d := range drives {
		if d.Path == path {
			return true
		}
	}
	return false
}
