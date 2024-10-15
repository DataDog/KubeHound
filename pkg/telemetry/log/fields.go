package log

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	FieldK8sTypeKey          = "k8s_type"
	FieldCountKey            = "count"
	FieldNodeTypeKey         = "node_type"
	FieldVertexTypeKey       = "vertex_type"
	FieldClusterKey          = "cluster"
	FieldComponentKey        = "component"
	FieldRunIDKey            = "run_id"
	FieldTeamKey             = "team"
	FieldServiceKey          = "service"
	FieldIngestorPipelineKey = "ingestor_pipeline"
	FieldDumpPipelineKey     = "dump_pipeline"
	FieldPathKey             = "path"
	FieldEntityKey           = "entity"
)

type contextKey int

const (
	ContextFieldRunID contextKey = iota
	ContextFieldCluster
	ContextFieldComponent
)

func convertField(value any) string {
	val, err := value.(string)
	if !err {
		return ""
	}
	return val
}

func SpanSetDefaultField(ctx context.Context, span ddtrace.Span) {
	runID := convertField(ctx.Value(ContextFieldRunID))
	if runID != "" {
		span.SetTag(FieldRunIDKey, convertField(runID))
	}

	cluster := convertField(ctx.Value(ContextFieldCluster))
	if cluster != "" {
		span.SetTag(FieldClusterKey, convertField(cluster))
	}
}

func FieldK8sType(k8sType string) string {
	return fmt.Sprintf("%s:%s", FieldK8sTypeKey, k8sType)
}

func FieldCount(count int) string {
	return fmt.Sprintf("%s:%d", FieldCountKey, count)
}

func FieldNodeType(nodeType string) string {
	return fmt.Sprintf("%s:%s", FieldNodeTypeKey, nodeType)
}

func FieldVertexType(vertexType string) string {
	return fmt.Sprintf("%s:%s", FieldVertexTypeKey, vertexType)
}

func FieldCluster(cluster string) string {
	return fmt.Sprintf("%s:%s", FieldClusterKey, cluster)
}

func FieldComponent(component string) string {
	return fmt.Sprintf("%s:%s", FieldComponentKey, component)
}

func FieldRunID(runID string) string {
	return fmt.Sprintf("%s:%s", FieldRunIDKey, runID)
}

func FieldTeam(team string) string {
	return fmt.Sprintf("%s:%s", FieldTeamKey, team)
}

func FieldService(service string) string {
	return fmt.Sprintf("%s:%s", FieldServiceKey, service)
}

func FieldIngestorPipeline(ingestorPipeline string) string {
	return fmt.Sprintf("%s:%s", FieldIngestorPipelineKey, ingestorPipeline)
}

func FieldDumpPipeline(dumpPipeline string) string {
	return fmt.Sprintf("%s:%s", FieldDumpPipelineKey, dumpPipeline)
}

// Field aliased here to make it easier to adopt this package
type Field = zapcore.Field

type msec time.Duration

func (f msec) String() string {
	ms := time.Duration(f) / time.Millisecond
	if ms < 10 {
		us := time.Duration(f) / time.Microsecond
		return fmt.Sprintf("%0.1fms", float64(us)/1000.0)
	}
	return strconv.Itoa(int(ms)) + "ms"
}

// Msec writes a duration in milliseconds.  If the time is <10msec, it is
// given to one decimal point.
func Msec(key string, dur time.Duration) Field {
	return zap.Stringer(key, msec(dur))
}

// Duration using its standard String() representation.
func Duration(key string, value time.Duration) Field {
	return zap.Duration(key, value)
}

type floatFmt struct {
	value  float64
	format string
}

func (f floatFmt) String() string {
	return fmt.Sprintf(f.format, f.value)
}

// Float writes a float64 value using the printf-style fmt string.
func Float(key string, val float64, fmt string) Field {
	return zap.Stringer(key, floatFmt{val, fmt})
}

// Base64 writes value encoded as base64.
func Base64(key string, value []byte) Field {
	return zap.Binary(key, value)
}

// Float64 writes a float value.
func Float64(key string, val float64) Field {
	return zap.Float64(key, val)
}

// Bool writes "true" or "false" for value.
func Bool(key string, value bool) Field {
	return zap.Bool(key, value)
}

// Dur writes a duration field truncated to the given duration.
func Dur(key string, value time.Duration, truncate ...time.Duration) Field {
	trunc := time.Microsecond
	if len(truncate) > 0 {
		trunc = truncate[0]
	}
	return zap.Duration(key, value-(value%trunc))
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type richError struct {
	kind          string
	message       string
	stack         errors.StackTrace
	handlingStack errors.StackTrace
}

func (f richError) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if f.kind != "" {
		enc.AddString("kind", f.kind)
	}
	if f.message != "" {
		enc.AddString("message", f.message)
	}
	marshalStacktrace(enc, "stack", f.stack)
	marshalStacktrace(enc, "handling_stack", f.handlingStack)
	return nil
}

func marshalStacktrace(enc zapcore.ObjectEncoder, fieldName string, st errors.StackTrace) {
	if st != nil {
		s := fmt.Sprintf("%+v", st)
		if len(s) > 0 && s[0] == '\n' {
			s = s[1:]
		}
		enc.AddString(fieldName, s)
	}
}

// RichError writes an error in the standard format expected by Datadog:
//
//   - type of error in `error.kind`
//   - `err.Error()` in `error.message`
//   - stack trace from the first error that has one in the chain of wrapped errors starting from err in `error.stack`, or [RichError] caller stack trace if no such stack trace was found.
//   - [RichError] caller stack trace in `error.handling_stack` if a stack trace was found from err.
func RichError(err error) Field {
	if err == nil {
		return zap.Skip()
	}

	var callerStackTrace errors.StackTrace
	if errWithStackTrace, ok := errors.WithStack(err).(stackTracer); ok {
		callerStackTrace = errWithStackTrace.StackTrace()
		if len(callerStackTrace) > 0 {
			callerStackTrace = callerStackTrace[1:]
		}
	}

	re := richError{kind: reflect.TypeOf(err).String(), message: err.Error()}

	if stacktrace := getStacktrace(err); stacktrace != nil {
		re.stack = stacktrace
		re.handlingStack = callerStackTrace
	} else {
		re.stack = callerStackTrace
	}

	return zap.Object("error", re)
}

// Interface to unwrap joined errors.Join https://pkg.go.dev/errors#Join
type UnwrapJoin interface {
	Unwrap() []error
}

// Interface to unwrap joined multierror.Append https://pkg.go.dev/github.com/hashicorp/go-multierror#Error.WrappedErrors
type UnwrapMultierror interface {
	WrappedErrors() []error
}

func getStacktrace(err error) errors.StackTrace {
	errorsToTest := []error{err}

	for index := 0; index < len(errorsToTest); index++ {
		testedErr := errorsToTest[index]

		if stackTracer, ok := testedErr.(stackTracer); ok {
			return stackTracer.StackTrace()
		}

		if joinErr, ok := testedErr.(UnwrapJoin); ok {
			errorsToTest = append(errorsToTest, joinErr.Unwrap()...)
		} else if joinErr, ok := testedErr.(UnwrapMultierror); ok {
			errorsToTest = append(errorsToTest, joinErr.WrappedErrors()...)
		} else if unwrapped := errors.Unwrap(testedErr); unwrapped != nil {
			errorsToTest = append(errorsToTest, unwrapped)
		}
	}

	return nil
}

// ErrorField writes an error.
func ErrorField(err error) Field {
	if err == nil {
		return zap.Skip()
	}
	return zap.String("error", err.Error())
}

// ErrorWithStackField writes an error. Prints message and stack if available.
func ErrorWithStackField(err error) Field {
	if err == nil {
		return zap.Skip()
	}
	return Object("error", err)
}

// NamedError writes an error with a custom name.
func NamedError(key string, err error) Field {
	return zap.NamedError(key, err)
}

// Float32 writes a float32 value.
func Float32(key string, value float32) Field {
	return zap.Float64(key, float64(value))
}

// Int writes an int value.
func Int(key string, value int) Field {
	return zap.Int64(key, int64(value))
}

// Ints writes an int slice as an array.
func Ints(key string, value []int) Field {
	return zap.Ints(key, value)
}

// Int64 writes an int64 value.
func Int64(key string, value int64) Field {
	return zap.Int64(key, value)
}

// Int64s writes an int64 slice as an array.
func Int64s(key string, value []int64) Field {
	return zap.Int64s(key, value)
}

// Int32 writes an int32 value.
func Int32(key string, value int32) Field {
	return zap.Int64(key, int64(value))
}

// Int32s writes a slice of int32s.
func Int32s(key string, value []int32) Field {
	return zap.Int32s(key, value)
}

type objectField struct {
	o interface{}
}

func (o objectField) String() string {
	return fmt.Sprintf("%+v", o.o)
}

// Object writes an object with "%+v".
func Object(key string, value interface{}) Field {
	return zap.Stringer(key, objectField{value})
}

// StructuredObject adds value as a structured object.
// value must implement zap.MarshalLogObject. Examples of such implementations can be found here:
// https://github.com/uber-go/zap/blob/9b86a50a3e27e0e12ccb6b47288de27df7fd3d5b/example_test.go#L176-L186
func StructuredObject(key string, value zapcore.ObjectMarshaler) Field {
	return zap.Object(key, value)
}

// String writes a string value.
func String(key, value string) Field {
	return zap.String(key, value)
}

// Strings writes a slice of strings.
func Strings(key string, value []string) Field {
	return zap.Strings(key, value)
}

// Stringer writes the output of the value's String method.
// The Stringer's String method is called lazily.
func Stringer(key string, value fmt.Stringer) Field {
	return zap.Stringer(key, value)
}

// Stringers writes the output of the value's String methods.
// The Stringer's String methods are called lazily.
func Stringers[T fmt.Stringer](key string, value []T) Field {
	return zap.Stringers(key, value)
}

// Byte writes a single byte as its ascii representation.
func Byte(key string, value byte) Field {
	return zap.ByteString(key, []byte{value})
}

// Bytes writes the []bytes as a string, up to limit characters.
func Bytes(key string, value []byte, limit int) Field {
	if limit > 0 && limit < len(value) {
		return zap.ByteString(key, value[:limit])
	}
	return zap.ByteString(key, value)
}

// stringfStringer implements the Stringf field with lazy evaluation.
type stringfStringer struct {
	format string
	args   []interface{}
}

func (s *stringfStringer) String() string {
	return fmt.Sprintf(s.format, s.args...)
}

// Stringf writes fmt.Sprintf(format, args...). It is evaluated lazily,
// only if the log message is going to be emitted.
func Stringf(key, format string, args ...interface{}) Field {
	return zap.Stringer(key, &stringfStringer{format, args})
}

// Time writes the value as a Unix timestamp.
func Time(key string, value time.Time) Field {
	return Int64(key, value.Unix())
}

// Uint writes a uint.
func Uint(key string, value uint) Field {
	return zap.Uint64(key, uint64(value))
}

// Uints writes a uint slice as an array.
func Uints(key string, value []uint) Field {
	return zap.Uints(key, value)
}

// Uint64 writes a uint64.
func Uint64(key string, value uint64) Field {
	return zap.Uint64(key, value)
}

// Uint64s writes a uint64 slice as an array.
func Uint64s(key string, value []uint64) Field {
	return zap.Uint64s(key, value)
}

// Uint32 writes a uint32.
func Uint32(key string, value uint32) Field {
	return zap.Uint64(key, uint64(value))
}

// Uint32s writes a uint32 slice as an array.
func Uint32s(key string, value []uint32) Field {
	return zap.Uint32s(key, value)
}

// Skip returns a no-op field
func Skip() Field {
	return zap.Skip()
}

type pct struct {
	part, whole float64
}

func (p pct) String() string {
	return fmt.Sprintf("%0.3f%%", (p.part/p.whole)*100)
}

// Percent writes out a percent out of 100%.
func Percent(key string, part, whole float64) Field {
	return zap.Stringer(key, pct{part, whole})
}

// PercentInt writes out a percent out of 100%.
func PercentInt(key string, part, whole int) Field {
	return zap.Stringer(key, pct{float64(part), float64(whole)})
}
