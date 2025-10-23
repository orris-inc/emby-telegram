package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[37m"
)

type colorHandler struct {
	writer io.Writer
	opts   *slog.HandlerOptions
	attrs  []slog.Attr
	groups []string
}

func newColorHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &colorHandler{
		writer: w,
		opts:   opts,
	}
}

func (h *colorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *colorHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := make([]byte, 0, 1024)

	buf = append(buf, colorGray...)
	buf = r.Time.AppendFormat(buf, "2006-01-02 15:04:05")
	buf = append(buf, colorReset...)
	buf = append(buf, ' ')

	levelColor := h.levelColor(r.Level)
	buf = append(buf, levelColor...)
	buf = append(buf, r.Level.String()...)
	buf = append(buf, colorReset...)
	buf = append(buf, ' ')

	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		buf = append(buf, colorGray...)
		buf = append(buf, filepath.Base(f.File)...)
		buf = append(buf, ':')
		buf = append(buf, fmt.Sprint(f.Line)...)
		buf = append(buf, colorReset...)
		buf = append(buf, ' ')
	}

	buf = append(buf, r.Message...)

	r.Attrs(func(a slog.Attr) bool {
		buf = append(buf, ' ')
		buf = append(buf, a.Key...)
		buf = append(buf, '=')
		buf = append(buf, a.Value.String()...)
		return true
	})

	buf = append(buf, '\n')

	_, err := h.writer.Write(buf)
	return err
}

func (h *colorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &colorHandler{
		writer: h.writer,
		opts:   h.opts,
		attrs:  append(h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *colorHandler) WithGroup(name string) slog.Handler {
	return &colorHandler{
		writer: h.writer,
		opts:   h.opts,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}

func (h *colorHandler) levelColor(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return colorBlue
	case slog.LevelInfo:
		return colorGreen
	case slog.LevelWarn:
		return colorYellow
	case slog.LevelError:
		return colorRed
	default:
		return colorReset
	}
}
