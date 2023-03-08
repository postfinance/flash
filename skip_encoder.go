package flash

import (
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type skipEncoder struct {
	zapcore.Encoder
	Skip map[string]struct{}
}

type newEncoderFunc func(cfg zapcore.EncoderConfig) zapcore.Encoder

//nolint:gocritic // we do not care about hugeparam critic, since it is only used on creation time
func newSkipEncoder(newEncoder newEncoderFunc, cfg zapcore.EncoderConfig, skip ...string) zapcore.Encoder {
	m := make(map[string]struct{}, len(skip))

	for _, s := range skip {
		m[s] = struct{}{}
	}

	return &skipEncoder{
		Encoder: newEncoder(cfg),
		Skip:    m,
	}
}

// EncodeEntry implements the zap encode entry interface.
//
//nolint:gocritic // we don't care about the hugeparam critic, because the interface is like that
func (e *skipEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	filtered := make([]zapcore.Field, 0, len(fields))

	for _, field := range fields {
		if _, ok := e.Skip[field.Key]; ok {
			continue
		}

		filtered = append(filtered, field)
	}

	return e.Encoder.EncodeEntry(entry, filtered)
}
