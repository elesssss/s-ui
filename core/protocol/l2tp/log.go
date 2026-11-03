package l2tp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/sagernet/sing-box/log"
)

// protocolLogHandler routes veepin's control-plane diagnostics through the
// panel's logger. A nil veepin logger silently discards negotiation failures.
type protocolLogHandler struct {
	logger log.ContextLogger
	attrs  string
	group  string
}

func (h protocolLogHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h protocolLogHandler) Handle(ctx context.Context, r slog.Record) error {
	var fields strings.Builder
	fields.WriteString(h.attrs)
	r.Attrs(func(a slog.Attr) bool { appendLogAttr(&fields, h.group, a); return true })
	message := r.Message + fields.String()
	// The destination logger applies the configured panel log level.
	switch {
	case r.Level >= slog.LevelError:
		h.logger.ErrorContext(ctx, message)
	case r.Level >= slog.LevelWarn:
		h.logger.WarnContext(ctx, message)
	case r.Level >= slog.LevelInfo:
		h.logger.InfoContext(ctx, message)
	default:
		h.logger.DebugContext(ctx, message)
	}
	return nil
}

func (h protocolLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var fields strings.Builder
	fields.WriteString(h.attrs)
	for _, a := range attrs {
		appendLogAttr(&fields, h.group, a)
	}
	h.attrs = fields.String()
	return h
}

func (h protocolLogHandler) WithGroup(name string) slog.Handler {
	if name != "" {
		h.group += name + "."
	}
	return h
}

func appendLogAttr(out *strings.Builder, group string, a slog.Attr) {
	a.Value = a.Value.Resolve()
	if a.Equal(slog.Attr{}) {
		return
	}
	if a.Value.Kind() == slog.KindGroup {
		if a.Key != "" {
			group += a.Key + "."
		}
		for _, child := range a.Value.Group() {
			appendLogAttr(out, group, child)
		}
		return
	}
	fmt.Fprintf(out, " %s%s=%q", group, a.Key, a.Value.String())
}
