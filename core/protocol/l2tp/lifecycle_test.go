//go:build linux && with_gvisor

package l2tp

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/log"
)

func TestWorkerExitHealthAndCleanup(t *testing.T) {
	for _, cause := range []error{errors.New("test read failure"), nil} {
		in := &Inbound{logger: log.NewNOPFactory().Logger()}
		cleaned := false
		in.runServer(func() error { return cause }, func() error { cleaned = true; return nil })
		if in.HealthError() == nil {
			t.Fatal("dead worker reports healthy")
		}
		if cause != nil && !errors.Is(in.HealthError(), cause) {
			t.Fatal("read failure was lost")
		}
		if !cleaned {
			t.Fatal("failed worker retains its sockets")
		}
		_ = in.Close()
	}
}

func TestAdministrativeCloseIsHealthy(t *testing.T) {
	in := &Inbound{logger: log.NewNOPFactory().Logger()}
	if err := in.Close(); err != nil {
		t.Fatal(err)
	}
	in.runServer(func() error { return nil }, func() error { t.Fatal("double-close on ordinary shutdown"); return nil })
	if in.HealthError() != nil {
		t.Fatal("administrative stop reported as a failure")
	}
}

func TestEmptyInboundRemainsHealthy(t *testing.T) {
	raw, err := newInbound(context.Background(), nil, log.NewNOPFactory().Logger(), "empty", Options{})
	if err != nil {
		t.Fatal(err)
	}
	in := raw.(*Inbound)
	if err := in.Start(adapter.StartStatePostStart, adapter.NewScope(context.Background(), log.NewNOPFactory().Logger())); err != nil {
		t.Fatal(err)
	}
	if in.HealthError() != nil {
		t.Fatal(in.HealthError())
	}
	_ = in.Close()
}

// sing-box 1.14.4 removes inbounds by closing their scopes rather than calling
// Close directly. Even an empty inbound must register its cleanup there.
func TestScopeCloseClosesInbound(t *testing.T) {
	scope := adapter.NewScope(context.Background(), log.NewNOPFactory().Logger())
	raw, err := newInbound(context.Background(), nil, log.NewNOPFactory().Logger(), "empty", Options{})
	if err != nil {
		t.Fatal(err)
	}
	in := raw.(*Inbound)
	if err := in.Start(adapter.StartStateInitialize, scope); err != nil {
		t.Fatal(err)
	}
	if err := scope.Close(); err != nil {
		t.Fatal(err)
	}
	if err := in.Start(adapter.StartStatePostStart, scope); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("scope removal left the inbound usable: %v", err)
	}
}
