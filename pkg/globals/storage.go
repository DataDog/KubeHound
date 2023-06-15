package globals

import "time"

const (
	DefaultRetry             int           = 10               // number of tries before failing
	DefaultRetryDelay        time.Duration = 10 * time.Second // in second
	DefaultConnectionTimeout time.Duration = 5 * time.Second  // in second
)
