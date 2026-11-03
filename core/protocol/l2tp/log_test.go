package l2tp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/sagernet/sing-box/log"
)

type protocolLogCapture struct {
	log.ContextLogger
	level, message string
}

func (l *protocolLogCapture) InfoContext(_ context.Context, args ...any) {
	l.level, l.message = "info", fmt.Sprint(args...)
}
func (l *protocolLogCapture) WarnContext(_ context.Context, args ...any) {
	l.level, l.message = "warn", fmt.Sprint(args...)
}
func (l *protocolLogCapture) ErrorContext(_ context.Context, args ...any) {
	l.level, l.message = "error", fmt.Sprint(args...)
}
func (l *protocolLogCapture) DebugContext(_ context.Context, args ...any) {
	l.level, l.message = "debug", fmt.Sprint(args...)
}

func TestProtocolLogsReachPanelWithLevelAndContext(t *testing.T) {
	capture := &protocolLogCapture{}
	base := slog.New(protocolLogHandler{logger: capture})
	child := base.With("peer", "test").WithGroup("ike").With("exchange", "main")
	for _, tc := range []struct {
		level slog.Level
		want  string
	}{{slog.LevelDebug, "debug"}, {slog.LevelInfo, "info"}, {slog.LevelWarn, "warn"}, {slog.LevelError, "error"}} {
		child.Log(context.Background(), tc.level, "NO-PROPOSAL-CHOSEN", slog.Group("detail", "step", 2))
		if capture.level != tc.want || !strings.Contains(capture.message, "NO-PROPOSAL-CHOSEN") || !strings.Contains(capture.message, `ike.detail.step="2"`) || !strings.Contains(capture.message, `peer="test"`) {
			t.Fatalf("log lost level or fields: %s %s", capture.level, capture.message)
		}
	}
	base.Info("plain")
	if capture.message != "plain" {
		t.Fatal("child attributes mutated parent handler")
	}
}
