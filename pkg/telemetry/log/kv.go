package log

import (
	"encoding/base64"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var (
	DefaultRemovedFields = []string{FieldTeamKey, FieldServiceKey, FieldRunIDKey, FieldClusterKey, FieldComponentKey, spanIDKey, traceIDKey}
	bufferpool           = buffer.NewPool()
)

type kvEncoder struct {
	*zapcore.EncoderConfig
	buf *buffer.Buffer
}

// NewKeyValueEncoder creates a key/value encoder that emits logs with a very basic
// "key=value" formatting.
func NewKeyValueEncoder(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
	return newkvEncoder(cfg), nil
}

func newkvEncoder(cfg zapcore.EncoderConfig) *kvEncoder {
	return &kvEncoder{
		EncoderConfig: &cfg,
		buf:           bufferpool.Get(),
	}
}

// DumpBuffer dumps this encoder's buffer as a string, resetting it.
// Useful for testing.
func (enc *kvEncoder) DumpBuffer() string {
	defer enc.buf.Reset()
	return enc.buf.String()
}

// forceElementSeparator even if it looks like we're adding a value;
// useful for functions that know they are appending a key
func (enc *kvEncoder) forceElementSeparator() {
	last := enc.buf.Len() - 1
	if last < 0 {
		return
	}
	bb := enc.buf.Bytes()
	switch bb[last] {
	case ' ', '(', '{':
		return
	}
	enc.buf.AppendByte(' ')
}

func (enc *kvEncoder) addElementSeparator() {
	last := enc.buf.Len() - 1
	if last < 0 {
		return
	}
	// XXX(jason): technically, this means values that end in "=" or
	// "[" will not get a separator;  this may not matter.
	bb := enc.buf.Bytes()
	switch bb[last] {
	case ' ', '=', '(', '{':
		return
	}
	enc.buf.AppendByte(' ')
}

func (enc *kvEncoder) addKey(key string) {
	enc.forceElementSeparator()
	// if the key is left out, we might be part of a "compound" nested object that
	// is supposed to be flat, so elide the key and the "=" sign
	if len(key) > 0 {
		enc.buf.AppendString(key)
		enc.buf.AppendByte('=')
	}
}

func (enc *kvEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	enc.addKey(key)
	return enc.AppendArray(marshaler)
}

func (enc *kvEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	enc.addKey(key)
	return marshaler.MarshalLogObject(enc)
}

func (enc *kvEncoder) AddBinary(key string, value []byte) {
	enc.AddString(key, base64.StdEncoding.EncodeToString(value))
}

func (enc *kvEncoder) AddByteString(key string, value []byte) {
	enc.addKey(key)
	enc.buf.Write(value)
}
func (enc *kvEncoder) AddBool(key string, value bool) {
	enc.addKey(key)
	enc.AppendBool(value)
}

func (enc *kvEncoder) AddComplex128(key string, value complex128) {
	enc.addKey(key)
	enc.AppendComplex128(value)
}

func (enc *kvEncoder) AddComplex64(key string, value complex64) {
	enc.AddComplex128(key, complex128(value))
}

func (enc *kvEncoder) AddDuration(key string, value time.Duration) {
	enc.AddString(key, value.String())
}

func (enc *kvEncoder) AddFloat64(key string, value float64) {
	enc.addKey(key)
	enc.AppendFloat64(value)
}
func (enc *kvEncoder) AddFloat32(key string, value float32) { enc.AddFloat64(key, float64(value)) }

func (enc *kvEncoder) AddInt64(key string, value int64) {
	enc.addKey(key)
	enc.buf.AppendString(strconv.FormatInt(value, 10))
}
func (enc *kvEncoder) AddInt(key string, value int)     { enc.AddInt64(key, int64(value)) }
func (enc *kvEncoder) AddInt32(key string, value int32) { enc.AddInt64(key, int64(value)) }
func (enc *kvEncoder) AddInt16(key string, value int16) { enc.AddInt64(key, int64(value)) }
func (enc *kvEncoder) AddInt8(key string, value int8)   { enc.AddInt64(key, int64(value)) }

func (enc *kvEncoder) AddString(key, value string) {
	enc.addKey(key)
	// If we have spaces, we surround the value with double quotes and escape existing double quotes
	if strings.Contains(value, " ") {
		value = `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}
	enc.buf.Write([]byte(value))
}
func (enc *kvEncoder) AddRawString(key, value string) {
	enc.addKey(key)
	enc.buf.Write([]byte(value))
}

func (enc *kvEncoder) AddTime(key string, value time.Time) {
	enc.addKey(key)
	enc.AppendTime(value)
}
func (enc *kvEncoder) AddUint64(key string, value uint64) {
	enc.addKey(key)
	enc.buf.AppendString(strconv.FormatUint(value, 10))
}
func (enc *kvEncoder) AddUint(key string, value uint)       { enc.AddUint64(key, uint64(value)) }
func (enc *kvEncoder) AddUint32(key string, value uint32)   { enc.AddUint64(key, uint64(value)) }
func (enc *kvEncoder) AddUint16(key string, value uint16)   { enc.AddUint64(key, uint64(value)) }
func (enc *kvEncoder) AddUint8(key string, value uint8)     { enc.AddUint64(key, uint64(value)) }
func (enc *kvEncoder) AddUintptr(key string, value uintptr) { enc.AddUint64(key, uint64(value)) }
func (enc *kvEncoder) AddReflected(key string, value interface{}) error {
	enc.AddRawString(key, fmt.Sprintf("%v", value))
	return nil
}

func (enc *kvEncoder) OpenNamespace(key string) {}

func (enc *kvEncoder) AppendObject(marshaler zapcore.ObjectMarshaler) error {
	return marshaler.MarshalLogObject(enc)
}

func (enc *kvEncoder) AppendReflected(val interface{}) error {
	enc.AppendString(fmt.Sprintf("%v", val))
	return nil
}

// Implement zapcore.PrimitiveArrayEncoder
func (enc *kvEncoder) AppendBool(value bool) {
	enc.addElementSeparator()
	enc.buf.AppendBool(value)
}

func (enc *kvEncoder) AppendByteString(value []byte) {
	enc.addElementSeparator()
	enc.buf.Write(value)
}

func (enc *kvEncoder) AppendDuration(value time.Duration) {
	enc.AppendString(value.String())
}

func (enc *kvEncoder) AppendComplex128(value complex128) {
	enc.addElementSeparator()
	r, i := float64(real(value)), float64(imag(value))
	enc.buf.AppendFloat(r, 64)
	enc.buf.AppendByte('+')
	enc.buf.AppendFloat(i, 64)
	enc.buf.AppendByte('i')
}

func (enc *kvEncoder) AppendArray(marshaler zapcore.ArrayMarshaler) error {
	enc.addElementSeparator()
	enc.buf.AppendByte('{')
	err := marshaler.MarshalLogArray(enc)
	enc.buf.AppendByte('}')
	return err
}
func (enc *kvEncoder) AppendComplex64(value complex64) { enc.AppendComplex128(complex128(value)) }
func (enc *kvEncoder) AppendFloat64(value float64) {
	enc.addElementSeparator()
	enc.buf.AppendString(strconv.FormatFloat(value, 'g', -1, 64))
}
func (enc *kvEncoder) AppendFloat32(value float32) { enc.AppendFloat64(float64(value)) }
func (enc *kvEncoder) AppendInt64(value int64)     { enc.AppendString(strconv.FormatInt(value, 10)) }
func (enc *kvEncoder) AppendInt(value int)         { enc.AppendInt64(int64(value)) }
func (enc *kvEncoder) AppendInt32(value int32)     { enc.AppendInt64(int64(value)) }
func (enc *kvEncoder) AppendInt16(value int16)     { enc.AppendInt64(int64(value)) }
func (enc *kvEncoder) AppendInt8(value int8)       { enc.AppendInt64(int64(value)) }
func (enc *kvEncoder) AppendString(value string) {
	enc.addElementSeparator()
	enc.buf.AppendString(value)
}
func (enc *kvEncoder) AppendTime(value time.Time) {
	enc.addElementSeparator()
	if enc.EncodeTime != nil {
		enc.EncodeTime(value, enc)
	} else {
		enc.AppendString(fmt.Sprint(value))
	}
}
func (enc *kvEncoder) AppendUint64(value uint64)   { enc.AppendString(strconv.FormatUint(value, 10)) }
func (enc *kvEncoder) AppendUint(value uint)       { enc.AppendUint64(uint64(value)) }
func (enc *kvEncoder) AppendUint32(value uint32)   { enc.AppendUint64(uint64(value)) }
func (enc *kvEncoder) AppendUint16(value uint16)   { enc.AppendUint64(uint64(value)) }
func (enc *kvEncoder) AppendUint8(value uint8)     { enc.AppendUint64(uint64(value)) }
func (enc *kvEncoder) AppendUintptr(value uintptr) { enc.AppendUint64(uint64(value)) }

func (enc *kvEncoder) Clone() zapcore.Encoder {
	clone := enc.clone()
	clone.buf.Write(enc.buf.Bytes())
	return clone
}

func (enc *kvEncoder) clone() *kvEncoder {
	clone := newkvEncoder(*enc.EncoderConfig)
	clone.buf = bufferpool.Get()
	return clone
}

func (enc *kvEncoder) EncodeEntry(ent zapcore.Entry, rawfields []zapcore.Field) (*buffer.Buffer, error) {
	c := enc.clone()

	fields := make([]zapcore.Field, 0, len(rawfields))
	for _, field := range rawfields {
		if slices.Contains(DefaultRemovedFields, field.Key) {
			continue
		}
		fields = append(fields, field)
	}
	// we overload the time encoder to determine whether or not we will also append
	// a hostname and appname.  This way, we can avoid doing so in environments where
	// this is prepended to us by some other means.
	if c.TimeKey != "" && c.EncodeTime != nil {
		c.EncodeTime(ent.Time, c)
	}

	if c.LevelKey != "" && c.EncodeLevel != nil {
		c.EncodeLevel(ent.Level, c)
	}

	if ent.LoggerName != "" && c.NameKey != "" {
		nameEncoder := c.EncodeName
		if nameEncoder == nil {
			nameEncoder = zapcore.FullNameEncoder
		}
		nameEncoder(ent.LoggerName, c)
	}

	if ent.Caller.Defined && c.CallerKey != "" && c.EncodeCaller != nil {
		c.AppendString("(") // we want to get the space here if needed
		c.EncodeCaller(ent.Caller, c)
		c.buf.AppendByte(')')

		// mimic seelog output and add dash between the preamble and the actual message
		c.AppendString("-")
	}

	// add the message even if MessageKey is not set
	c.AppendString(ent.Message)

	if enc.buf.Len() > 0 {
		c.addElementSeparator()
		c.buf.Write(enc.buf.Bytes())
	}

	for i := range fields {
		fields[i].AddTo(c)
	}

	if ent.Stack != "" && c.StacktraceKey != "" {
		c.buf.AppendByte('\n')
		c.buf.AppendString(ent.Stack)
	}

	// do not accidentally add a space between the line and the line ending
	if c.LineEnding != "" {
		c.buf.AppendString(c.LineEnding)
	} else {
		c.buf.AppendString(zapcore.DefaultLineEnding)
	}

	return c.buf, nil
}
