package cache

type writerOptions struct {
	Test            bool
	ExpectOverwrite bool
}

type WriterOption func(*writerOptions)

// Perform a test and set operation on writes.
// Only set the value if it does not currently exist. If the value does exist, return an ErrCacheEntryOverwrite
// error. Mutually exclusive with WithExpectedOverwrite.
func WithTest() WriterOption {
	return func(wOpts *writerOptions) {
		wOpts.Test = true
	}
}

// WithExpectedOverwrite signals that overwriting values is expected and suppresses logs & metrics generation.
// Mutually exclusive with WithTest.
func WithExpectedOverwrite() WriterOption {
	return func(wOpts *writerOptions) {
		wOpts.ExpectOverwrite = true
	}
}
