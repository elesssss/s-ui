package manager

import (
	"context"
	"errors"
	"testing"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/log"
)

type preparedProbe struct {
	initialized int
	closed      int
}

func (p *preparedProbe) Start(stage adapter.StartStage, scope *adapter.Scope) error {
	if stage == adapter.StartStateInitialize {
		p.initialized++
		scope.Add(func() error { p.closed++; return nil })
	}
	return nil
}

type preparedOutbound struct {
	adapter.Outbound
	*preparedProbe
}

func (*preparedOutbound) Type() string           { return "probe" }
func (*preparedOutbound) Tag() string            { return "one" }
func (*preparedOutbound) Dependencies() []string { return nil }

type preparedOutboundRegistry struct {
	adapter.OutboundRegistry
	probe *preparedProbe
}

func (r *preparedOutboundRegistry) CreateOutbound(_ context.Context, _ adapter.Router, _ log.ContextLogger, tag, _ string, _ any) (adapter.Outbound, error) {
	if tag == "fail" {
		return nil, errors.New("constructor failed")
	}
	return &preparedOutbound{preparedProbe: r.probe}, nil
}

type preparedTransport struct {
	adapter.DNSTransport
	*preparedProbe
}

func (p *preparedTransport) Start(stage adapter.StartStage, scope *adapter.Scope) error {
	return p.preparedProbe.Start(stage, scope)
}

func (*preparedTransport) Type() string           { return "probe" }
func (*preparedTransport) Tag() string            { return "one" }
func (*preparedTransport) Dependencies() []string { return nil }

type preparedTransportRegistry struct {
	adapter.DNSTransportRegistry
	probe *preparedProbe
}

func (r *preparedTransportRegistry) CreateDNSTransport(_ context.Context, _ log.ContextLogger, tag, _ string, _ any) (adapter.DNSTransport, error) {
	if tag == "fail" {
		return nil, errors.New("constructor failed")
	}
	return &preparedTransport{preparedProbe: r.probe}, nil
}

// A second constructor can fail before the manager's Initialize stage. Every
// earlier resource must still close, and later startup must not initialize it
// a second time or register its cleanup twice.
func TestPreparedManagerOwnsResourcesBeforeStartup(t *testing.T) {
	logger := log.NewNOPFactory().Logger()
	for _, protocol := range []string{"outbound", "dns"} {
		for _, startManager := range []bool{false, true} {
			t.Run(protocol+map[bool]string{false: "/construction-failure", true: "/initialized"}[startManager], func(t *testing.T) {
				ctx := context.Background()
				scope := adapter.NewScope(ctx, logger)
				defer scope.Close()
				probe := &preparedProbe{}
				var create func(string) error
				var lifecycle adapter.Lifecycle
				if protocol == "outbound" {
					m := NewOutboundManager(logger, &preparedOutboundRegistry{probe: probe}, nil, "one")
					m.Prepare(scope)
					create = func(tag string) error { return m.Create(ctx, nil, logger, tag, "probe", nil) }
					lifecycle = m
				} else {
					m := NewDNSTransportManager(logger, &preparedTransportRegistry{probe: probe}, nil, "one")
					m.Prepare(scope)
					create = func(tag string) error { return m.Create(ctx, logger, tag, "probe", nil) }
					lifecycle = m
				}
				if err := create("one"); err != nil {
					t.Fatal(err)
				}
				if err := create("fail"); err == nil {
					t.Fatal("constructor failure was swallowed")
				}
				if startManager {
					if err := lifecycle.Start(adapter.StartStateInitialize, scope); err != nil {
						t.Fatal(err)
					}
				}
				if probe.initialized != 1 {
					t.Fatalf("initialized %d times", probe.initialized)
				}
				if err := scope.Close(); err != nil {
					t.Fatal(err)
				}
				if probe.closed != 1 {
					t.Fatalf("closed %d times", probe.closed)
				}
			})
		}
	}
}
