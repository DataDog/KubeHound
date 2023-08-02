package cache

type writerOptions struct {
	Test            bool
	ExpectOverwrite bool
}

type WriterOption func(*writerOptions)

// Perform a test and set operation on writes. Only set the value if it does not currently exist. If the value does exist,
// return an ErrCacheEntryOverwrite error.
func WithTest() WriterOption {
	return func(wOpts *writerOptions) {
		wOpts.Test = true
	}
}

func WithExpectedOverwrite() WriterOption {
	return func(wOpts *writerOptions) {
		wOpts.ExpectOverwrite = true
	}
}
