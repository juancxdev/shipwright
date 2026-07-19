package harness

import "time"

func NowISOCompact() string {
	return time.Now().UTC().Format("20060102-150405")
}
