package graphdb

import "fmt"

// batchWriterError is an error type that wraps an error and indicates whether the
// error is retryable.
type batchWriterError struct {
	err       error
	retryable bool
}

func (e batchWriterError) Error() string {
	if e.err == nil {
		return fmt.Sprintf("batch writer error (retriable:%v)", e.retryable)
	}

	return fmt.Sprintf("batch writer error (retriable:%v): %v", e.retryable, e.err.Error())
}

func (e batchWriterError) Unwrap() error {
	return e.err
}
