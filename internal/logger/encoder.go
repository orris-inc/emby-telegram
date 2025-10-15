package logger

import (
	"strings"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const (
	callerWidth = 30
)

type alignedEncoder struct {
	zapcore.Encoder
	pool buffer.Pool
}

func newAlignedEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	return &alignedEncoder{
		Encoder: zapcore.NewConsoleEncoder(cfg),
		pool:    buffer.NewPool(),
	}
}

func (enc *alignedEncoder) Clone() zapcore.Encoder {
	return &alignedEncoder{
		Encoder: enc.Encoder.Clone(),
		pool:    enc.pool,
	}
}

func (enc *alignedEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf := enc.pool.Get()

	buf.AppendString(entry.Time.Format(time.RFC3339))
	buf.AppendByte('\t')

	buf.AppendString(entry.Level.CapitalString())
	buf.AppendByte('\t')

	caller := entry.Caller.TrimmedPath()
	if len(caller) > callerWidth {
		caller = "..." + caller[len(caller)-callerWidth+3:]
	}
	buf.AppendString(padRight(caller, callerWidth))
	buf.AppendByte('\t')

	buf.AppendString(entry.Message)

	buf.AppendByte('\n')

	return buf, nil
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
