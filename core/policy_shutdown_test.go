package core

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"sync"
	"testing"
	"time"

	mDNS "github.com/miekg/dns"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
)

func policyShutdownRuntime(t *testing.T) *policyRuntime {
	t.Helper()
	p := &policyRuntime{scope: adapter.NewScope(context.Background(), log.NewNOPFactory().Logger())}
	p.generation = &policyGeneration{runtime: p}
	probe := &policyCloseProbe{closed: make(chan struct{})}
	if err := p.scope.Start("probe", probe, adapter.StartStateInitialize); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.close() })
	return p
}

func TestPolicyStoppedAccessRejectsRequests(t *testing.T) {
	p := policyShutdownRuntime(t)
	r := &policyRouter{current: p.generation, all: []*policyGeneration{p.generation}}
	if err := r.shutdown(); err != nil {
		t.Fatal(err)
	}
	if g := r.acquire(); g != nil {
		g.release()
		t.Fatal("stopped router lent a closed policy")
	}
	// Native pointers are intentionally nil: none of these requests may reach
	// the native services after shutdown, even through a cached DNS handle.
	for _, d := range []adapter.DNSRouter{
		&accessDNS{policy: r},
		&policyDNS{generation: p.generation, policy: r},
	} {
		if _, err := d.Exchange(context.Background(), new(mDNS.Msg), adapter.DNSQueryOptions{}); !errors.Is(err, net.ErrClosed) {
			t.Fatalf("Exchange after shutdown: %v", err)
		}
		if _, err := d.Lookup(context.Background(), "reload.test", adapter.DNSQueryOptions{}); !errors.Is(err, net.ErrClosed) {
			t.Fatalf("Lookup after shutdown: %v", err)
		}
		called := make(chan error, 2)
		d.ExchangeAsync(context.Background(), new(mDNS.Msg), adapter.DNSQueryOptions{}, func(_ *mDNS.Msg, err error) { called <- err })
		select {
		case err := <-called:
			if !errors.Is(err, net.ErrClosed) {
				t.Fatalf("ExchangeAsync after shutdown: %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("ExchangeAsync dropped its callback after shutdown")
		}
		d.ClearCache()
		d.ResetNetwork()
		if domain, found := d.LookupReverseMapping(netip.MustParseAddr("192.0.2.1")); domain != "" || found {
			t.Fatal("stopped DNS returned a reverse mapping")
		}
	}
	if len(r.Rules()) != 0 || r.NeedFindProcess() || r.NeedFindNeighbor() || r.NeighborResolver() != nil {
		t.Fatal("stopped router exposed native policy services")
	}
	if _, found := r.RuleSet("test"); found {
		t.Fatal("stopped router exposed a rule-set")
	}
	r.ResetNetwork()
	r.PreMatch(adapter.InboundContext{}, nil)
	r.HijackDNSPacket(context.Background(), nil, nil, adapter.InboundContext{})
	client, server := net.Pipe()
	defer client.Close()
	if err := r.RouteConnection(context.Background(), server, adapter.InboundContext{}); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("RouteConnection after shutdown: %v", err)
	}
	if _, err := client.Write([]byte("x")); err == nil {
		t.Fatal("stopped router left a TCP connection open")
	}
	packet := &policyPacketProbe{closed: make(chan struct{})}
	if err := r.RoutePacketConnection(context.Background(), packet, adapter.InboundContext{}); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("RoutePacketConnection after shutdown: %v", err)
	}
	select {
	case <-packet.closed:
	default:
		t.Fatal("stopped router left a UDP connection open")
	}
	client, server = net.Pipe()
	defer client.Close()
	called := make(chan error, 2)
	r.RouteConnectionEx(context.Background(), server, adapter.InboundContext{}, func(err error) { called <- err })
	packet = &policyPacketProbe{closed: make(chan struct{})}
	r.RoutePacketConnectionEx(context.Background(), packet, adapter.InboundContext{}, func(err error) { called <- err })
	for i := 0; i < 2; i++ {
		select {
		case err := <-called:
			if !errors.Is(err, net.ErrClosed) {
				t.Fatalf("route callback after shutdown: %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("stopped route dropped its close callback")
		}
	}
	b := &Box{policy: r}
	b.CloseIdleConnections()
	if b.Outbound() != nil {
		t.Fatal("stopped box exposed an outbound manager")
	}
	if _, err := b.accessContext(); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("accessContext after shutdown: %v", err)
	}
	if p.generation.refs != 0 {
		t.Fatal("stopped requests retained a policy reference")
	}
}

func TestPolicyConcurrentShutdownAndPublication(t *testing.T) {
	for i := 0; i < 50; i++ {
		initial, candidate := policyShutdownRuntime(t), policyShutdownRuntime(t)
		r := &policyRouter{current: initial.generation, all: []*policyGeneration{initial.generation}}
		start := make(chan struct{})
		var workers sync.WaitGroup
		workers.Go(func() {
			<-start
			if err := r.shutdown(); err != nil {
				t.Error(err)
			}
		})
		workers.Go(func() {
			<-start
			if err := r.publish(candidate.generation, policyImpact{}); err != nil && !errors.Is(err, net.ErrClosed) {
				t.Error(err)
			}
		})
		close(start)
		workers.Wait()
		if !initial.closed.Load() || !candidate.closed.Load() || r.current != nil || len(r.all) != 0 {
			t.Fatal("concurrent publication resurrected or leaked a stopped policy")
		}
	}
}

func TestRetiredPolicyDNSNotificationsStayIsolated(t *testing.T) {
	old, current := policyShutdownRuntime(t), policyShutdownRuntime(t)
	r := &policyRouter{current: current.generation, all: []*policyGeneration{old.generation, current.generation}}
	t.Cleanup(func() { _ = r.shutdown() })
	old.generation.retire(policyImpact{})
	d := &policyDNS{generation: old.generation, policy: r}
	// A notification from an old network monitor may be delivered after a
	// DNS-only save. It must neither use the disposed DNS router nor redirect
	// that reset to the new policy's transports.
	d.ResetNetwork()
	d.ClearCache()
	if current.closed.Load() || current.generation.refs != 0 {
		t.Fatal("retired DNS notification affected the current generation")
	}
}

func TestPolicyCommitAfterShutdownReleasesCandidate(t *testing.T) {
	for _, prepared := range []bool{false, true} {
		t.Run(map[bool]string{false: "unchanged", true: "prepared"}[prepared], func(t *testing.T) {
			initial := policyShutdownRuntime(t)
			r := &policyRouter{current: initial.generation, all: []*policyGeneration{initial.generation}}
			before := option.Options{Route: &option.RouteOptions{Final: "old"}}
			b := &Box{policy: r, applied: before}
			u := &PolicyUpdate{box: b, options: option.Options{Route: &option.RouteOptions{Final: "new"}}}
			if prepared {
				u.runtime = policyShutdownRuntime(t)
			}
			if err := r.shutdown(); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				if err := u.Commit(); !errors.Is(err, net.ErrClosed) {
					t.Fatalf("Commit after shutdown: %v", err)
				}
			}
			u.Abort()
			if !sameOptions(before, b.applied) || r.current != nil || (prepared && !u.runtime.closed.Load()) {
				t.Fatal("late commit changed a stopped box or leaked its candidate")
			}
		})
	}
}
