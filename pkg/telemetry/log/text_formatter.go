package log

import (
	"time"

	logrus "github.com/sirupsen/logrus"
)

var (
	DefaultRemovedFields = []string{"component", "team", "service", "run_id"}
)

// FilteredTextFormatter is a logrus.TextFormatter that filters out some specific fields that are not needed for humans.
// These fields are usually more helpful for machines, and with json formatting for example.
type FilteredTextFormatter struct {
	tf            logrus.TextFormatter
	removedFields []string
}

func NewFilteredTextFormatter(removedFields []string) logrus.Formatter {
	return &FilteredTextFormatter{
		tf: logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.TimeOnly,
		},
		removedFields: removedFields,
	}
}

func (f *FilteredTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	for _, field := range f.removedFields {
		delete(entry.Data, field)
	}

	return f.tf.Format(entry)
}
