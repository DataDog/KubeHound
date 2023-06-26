package globals

import "time"

const (
	DefaultRetry      int           = 10 // number of tries before failing
	DefaultRetryDelay time.Duration = 10 * time.Second

	DefaultConnectionTimeout time.Duration = 5 * time.Second

	DefaultProfilerPeriod      time.Duration = 60 * time.Second
	DefaultProfilerCPUDuration time.Duration = 15 * time.Second
)
