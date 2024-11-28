package graphdb

import "fmt"

// errBatchWriter is an error type that wraps an error and indicates whether the
// error is retryable.
type errBatchWriter struct {
	err       error
	retryable bool
}

func (e errBatchWriter) Error() string {
	if e.err == nil {
		return fmt.Sprintf("batch writer error (retriable:%v)", e.retryable)
	}

	return fmt.Sprintf("batch writer error (retriable:%v): %v", e.retryable, e.err.Error())
}

func (e errBatchWriter) Unwrap() error {
	return e.err
}
