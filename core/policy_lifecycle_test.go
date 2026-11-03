package core

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/log"
	N "github.com/sagernet/sing/common/network"
)

type policyCloseProbe struct{ closed chan struct{} }

func (p *policyCloseProbe) Start(stage adapter.StartStage, scope *adapter.Scope) error {
	if stage == adapter.StartStateInitialize {
		scope.Add(p.Close)
	}
	return nil
}
func (p *policyCloseProbe) Close() error { close(p.closed); return nil }

type policyPacketProbe struct {
	N.PacketConn
	closed chan struct{}
}

func (p *policyPacketProbe) Close() error { close(p.closed); return nil }

func TestPolicyUDPInvalidationIsSelective(t *testing.T) {
	g := &policyGeneration{runtime: &policyRuntime{}}
	g.acquire()
	defer g.release()
	ctx := context.WithValue(context.Background(), policyContextKey{}, g)
	tracker := NewSessionTracker()
	a := &policyPacketProbe{closed: make(chan struct{})}
	b := &policyPacketProbe{closed: make(chan struct{})}
	first := tracker.RoutedPacketConnection(ctx, a, adapter.InboundContext{Inbound: "vpn", User: "alice"}, nil, &testOutbound{tag: "a"})
	second := tracker.RoutedPacketConnection(ctx, b, adapter.InboundContext{Inbound: "vpn", User: "bob"}, nil, &testOutbound{tag: "b"})
	defer first.Close()
	defer second.Close()
	g.retire(policyImpact{tags: map[string]bool{"a": true}})
	select {
	case <-a.closed:
	default:
		t.Fatal("affected UDP flow remained open")
	}
	select {
	case <-b.closed:
		t.Fatal("unrelated UDP flow was closed")
	default:
	}
	if sessions := tracker.Sessions(); len(sessions) != 1 || sessions[0].User != "bob" {
		t.Fatalf("UDP sessions: %v", sessions)
	}
}

func TestRetiredPolicyRejectsLateRoutesAndDrains(t *testing.T) {
	probe := &policyCloseProbe{closed: make(chan struct{})}
	p := &policyRuntime{scope: adapter.NewScope(context.Background(), log.NewNOPFactory().Logger())}
	if err := p.scope.Start("probe", probe, adapter.StartStateInitialize); err != nil {
		t.Fatal(err)
	}
	g := &policyGeneration{runtime: p}
	g.acquire() // an old route evaluation is still in flight
	g.retire(policyImpact{tags: map[string]bool{"a": true}})
	tracker := NewSessionTracker()
	client, server := net.Pipe()
	defer client.Close()
	ctx := context.WithValue(context.Background(), policyContextKey{}, g)
	tracked := tracker.RoutedConnection(ctx, server, adapter.InboundContext{Inbound: "vpn", User: "alice"}, nil, &testOutbound{tag: "a"})
	defer tracked.Close()
	if len(tracker.Sessions()) != 0 {
		t.Fatal("late route escaped invalidation")
	}
	select {
	case <-probe.closed:
		t.Fatal("disposed policy while route evaluation was in flight")
	default:
	}
	g.release()
	select {
	case <-probe.closed:
	case <-time.After(time.Second):
		t.Fatal("retired policy leaked")
	}
}

func TestPolicyConcurrentSessionPublicationAndRetirement(t *testing.T) {
	p := &policyRuntime{}
	g := &policyGeneration{runtime: p}
	g.acquire()
	tracker := NewSessionTracker()
	ctx := context.WithValue(context.Background(), policyContextKey{}, g)
	var workers sync.WaitGroup
	for i := 0; i < 100; i++ {
		workers.Go(func() {
			client, server := net.Pipe()
			defer client.Close()
			tracked := tracker.RoutedConnection(ctx, server, adapter.InboundContext{Inbound: "vpn"}, nil, &testOutbound{tag: "a"})
			defer tracked.Close()
		})
	}
	g.retire(policyImpact{all: true})
	workers.Wait()
	if len(tracker.Sessions()) != 0 {
		t.Fatal("partially published sessions survived retirement")
	}
	g.release()
}

func TestPolicyInvalidationSurvivesSubsequentDNSUpdates(t *testing.T) {
	p := &policyRuntime{}
	g := &policyGeneration{runtime: p}
	g.acquire()
	defer g.release()
	g.retire(policyImpact{tags: map[string]bool{"a": true}})
	for i := 0; i < 1000; i++ {
		g.retire(policyImpact{})
	}
	if !g.impact.matches("a") || g.impact.matches("b") || len(g.impact.tags) != 1 {
		t.Fatal("DNS updates forgot or expanded prior invalidation")
	}
}

func TestPolicyScopeRetirementKeepsAccessScope(t *testing.T) {
	logger := log.NewNOPFactory().Logger()
	access := adapter.NewScope(context.Background(), logger)
	defer access.Close()
	probe := &policyCloseProbe{closed: make(chan struct{})}
	p := &policyRuntime{scope: adapter.NewScope(access.Context(), logger)}
	if err := p.scope.Start("probe", probe, adapter.StartStateInitialize); err != nil {
		t.Fatal(err)
	}
	g := &policyGeneration{runtime: p}
	g.retire(policyImpact{})
	select {
	case <-probe.closed:
	default:
		t.Fatal("retired policy did not close its scoped resources")
	}
	if p.scope.Context().Err() == nil || access.Context().Err() != nil {
		t.Fatal("policy retirement leaked its scope or cancelled access")
	}
	if err := p.close(); err != nil {
		t.Fatal(err)
	}
}
