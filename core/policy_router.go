package core

import (
	"context"
	"errors"
	"io"
	"net"
	"net/netip"
	"sync"

	mDNS "github.com/miekg/dns"
	"github.com/sagernet/sing-box/adapter"
	N "github.com/sagernet/sing/common/network"
)

type policyContextKey struct{}

// Cumulative invalidation prevents a slow old route evaluation from surviving
// a later update. Maps avoid chains of predicates across repeated DNS saves.
type policyImpact struct {
	all  bool
	tags map[string]bool
}

func (p policyImpact) matches(tag string) bool { return p.all || p.tags[tag] }
func (p *policyImpact) merge(next policyImpact) {
	p.all = p.all || next.all
	if p.all {
		p.tags = nil
		return
	}
	if p.tags == nil && len(next.tags) > 0 {
		p.tags = make(map[string]bool)
	}
	for tag := range next.tags {
		p.tags[tag] = true
	}
}

// References cover route evaluation, asynchronous DNS and established traffic.
// No timer or polling is needed to reclaim an old policy.
type policyGeneration struct {
	mu      sync.Mutex
	refs    int
	retired bool
	impact  policyImpact
	flows   map[io.Closer]string
	runtime *policyRuntime
}

func (g *policyGeneration) acquire() { g.mu.Lock(); g.refs++; g.mu.Unlock() }
func (g *policyGeneration) release() {
	g.mu.Lock()
	g.refs--
	closeNow := g.retired && g.refs == 0
	g.mu.Unlock()
	if closeNow {
		go g.runtime.close()
	}
}

func (g *policyGeneration) track(conn io.Closer, outbound string) bool {
	g.mu.Lock()
	g.refs++
	if g.flows == nil {
		g.flows = make(map[io.Closer]string)
	}
	g.flows[conn] = outbound
	reject := g.retired && g.impact.matches(outbound)
	g.mu.Unlock()
	return reject
}

func (g *policyGeneration) untrack(conn io.Closer) {
	g.mu.Lock()
	delete(g.flows, conn)
	g.mu.Unlock()
	g.release()
}

func (g *policyGeneration) retire(impact policyImpact) {
	g.mu.Lock()
	g.retired = true
	g.impact.merge(impact)
	var closeFlows []io.Closer
	for conn, tag := range g.flows {
		if g.impact.matches(tag) {
			closeFlows = append(closeFlows, conn)
		}
	}
	closeNow := g.refs == 0
	g.mu.Unlock()
	for _, conn := range closeFlows {
		_ = conn.Close()
	}
	if closeNow {
		g.runtime.close()
	}
}

// Inbounds keep this identity for their entire lifetime. Policy publication is
// one pointer swap; it cannot close the IKE/L2TP/PPP server or its gVisor stack.
type policyRouter struct {
	adapter.Router // immutable initial router for lifecycle and non-routing APIs
	mu             sync.Mutex
	current        *policyGeneration
	all            []*policyGeneration
	stopped        bool
}

func (r *policyRouter) acquire() *policyGeneration {
	r.mu.Lock()
	defer r.mu.Unlock()
	g := r.current
	if r.stopped || g == nil {
		return nil
	}
	g.acquire()
	return g
}

func (r *policyRouter) publish(g *policyGeneration, impact policyImpact) error {
	r.mu.Lock()
	if r.stopped {
		r.mu.Unlock()
		_ = g.runtime.close()
		return net.ErrClosed
	}
	previous := append([]*policyGeneration(nil), r.all...)
	r.current = g
	r.all = append(r.all, g)
	r.mu.Unlock()
	// Also invalidate affected connections retained by earlier DNS-only updates.
	for _, old := range previous {
		old.retire(impact)
	}
	r.mu.Lock()
	live := r.all[:0]
	for _, old := range r.all {
		old.mu.Lock()
		keep := !old.runtime.closed.Load()
		old.mu.Unlock()
		if keep {
			live = append(live, old)
		}
	}
	clear(r.all[len(live):])
	r.all = live
	r.mu.Unlock()
	return nil
}

func (r *policyRouter) shutdown() error {
	r.mu.Lock()
	r.stopped = true
	r.current = nil
	all := append([]*policyGeneration(nil), r.all...)
	r.all = nil
	r.mu.Unlock()
	var err error
	for _, g := range all {
		g.retire(policyImpact{all: true})
		err = errors.Join(err, g.runtime.close())
	}
	return err
}

func (r *policyRouter) Rules() []adapter.Rule {
	g := r.acquire()
	if g == nil {
		return nil
	}
	defer g.release()
	return g.runtime.router.Rules()
}
func (r *policyRouter) RuleSet(tag string) (adapter.RuleSet, bool) {
	g := r.acquire()
	if g == nil {
		return nil, false
	}
	defer g.release()
	return g.runtime.router.RuleSet(tag)
}
func (r *policyRouter) NeedFindProcess() bool {
	g := r.acquire()
	if g == nil {
		return false
	}
	defer g.release()
	return g.runtime.router.NeedFindProcess()
}
func (r *policyRouter) NeedFindNeighbor() bool {
	g := r.acquire()
	if g == nil {
		return false
	}
	defer g.release()
	return g.runtime.router.NeedFindNeighbor()
}
func (r *policyRouter) NeighborResolver() adapter.NeighborResolver {
	g := r.acquire()
	if g == nil {
		return nil
	}
	defer g.release()
	return g.runtime.router.NeighborResolver()
}
func (r *policyRouter) ResetNetwork() {
	g := r.acquire()
	if g == nil {
		return
	}
	defer g.release()
	g.runtime.router.ResetNetwork()
}
func (r *policyRouter) PreMatch(m adapter.InboundContext, packet []byte) adapter.PreMatchResult {
	g := r.acquire()
	if g == nil {
		var result adapter.PreMatchResult
		return result
	}
	defer g.release()
	return g.runtime.router.PreMatch(m, packet)
}

func (r *policyRouter) RouteConnectionEx(ctx context.Context, conn net.Conn, m adapter.InboundContext, onClose N.CloseHandlerFunc) {
	g := r.acquire()
	if g == nil {
		N.CloseOnHandshakeFailure(conn, onClose, net.ErrClosed)
		return
	}
	defer g.release()
	reject := g.track(conn, "")
	defer g.untrack(conn)
	if reject {
		N.CloseOnHandshakeFailure(conn, onClose, net.ErrClosed)
		return
	}
	g.runtime.router.RouteConnectionEx(context.WithValue(ctx, policyContextKey{}, g), conn, m, onClose)
}
func (r *policyRouter) RoutePacketConnectionEx(ctx context.Context, conn N.PacketConn, m adapter.InboundContext, onClose N.CloseHandlerFunc) {
	g := r.acquire()
	if g == nil {
		N.CloseOnHandshakeFailure(conn, onClose, net.ErrClosed)
		return
	}
	defer g.release()
	reject := g.track(conn, "")
	defer g.untrack(conn)
	if reject {
		N.CloseOnHandshakeFailure(conn, onClose, net.ErrClosed)
		return
	}
	g.runtime.router.RoutePacketConnectionEx(context.WithValue(ctx, policyContextKey{}, g), conn, m, onClose)
}
func (r *policyRouter) RouteConnection(ctx context.Context, conn net.Conn, m adapter.InboundContext) error {
	g := r.acquire()
	if g == nil {
		_ = conn.Close()
		return net.ErrClosed
	}
	defer g.release()
	reject := g.track(conn, "")
	defer g.untrack(conn)
	if reject {
		_ = conn.Close()
		return net.ErrClosed
	}
	return g.runtime.router.RouteConnection(context.WithValue(ctx, policyContextKey{}, g), conn, m)
}
func (r *policyRouter) RoutePacketConnection(ctx context.Context, conn N.PacketConn, m adapter.InboundContext) error {
	g := r.acquire()
	if g == nil {
		_ = conn.Close()
		return net.ErrClosed
	}
	defer g.release()
	reject := g.track(conn, "")
	defer g.untrack(conn)
	if reject {
		_ = conn.Close()
		return net.ErrClosed
	}
	return g.runtime.router.RoutePacketConnection(context.WithValue(ctx, policyContextKey{}, g), conn, m)
}
func (r *policyRouter) HijackDNSPacket(ctx context.Context, payload []byte, writer N.PacketWriter, m adapter.InboundContext) {
	g := r.acquire()
	if g == nil {
		return
	}
	defer g.release()
	g.runtime.router.HijackDNSPacket(ctx, payload, writer, m)
}

// Native DNS routing is asynchronous. Retain its generation until the callback
// completes rather than disposing transports while a query is in flight.
type policyDNS struct {
	adapter.DNSRouter
	generation *policyGeneration
	policy     *policyRouter
}

func (d *policyDNS) acquire() bool {
	if d.policy != nil {
		d.policy.mu.Lock()
		defer d.policy.mu.Unlock()
		if d.policy.stopped {
			return false
		}
	}
	d.generation.mu.Lock()
	defer d.generation.mu.Unlock()
	if d.generation.retired && d.policy != nil {
		return false
	}
	d.generation.refs++
	return true
}

func (d *policyDNS) Exchange(ctx context.Context, message *mDNS.Msg, options adapter.DNSQueryOptions) (*mDNS.Msg, error) {
	if !d.acquire() {
		return (&accessDNS{policy: d.policy}).Exchange(ctx, message, options)
	}
	defer d.generation.release()
	return d.DNSRouter.Exchange(ctx, message, options)
}
func (d *policyDNS) Lookup(ctx context.Context, domain string, options adapter.DNSQueryOptions) ([]netip.Addr, error) {
	if !d.acquire() {
		return (&accessDNS{policy: d.policy}).Lookup(ctx, domain, options)
	}
	defer d.generation.release()
	return d.DNSRouter.Lookup(ctx, domain, options)
}

func (d *policyDNS) ExchangeAsync(ctx context.Context, message *mDNS.Msg, options adapter.DNSQueryOptions, callback func(*mDNS.Msg, error)) {
	if !d.acquire() {
		(&accessDNS{policy: d.policy}).ExchangeAsync(ctx, message, options, callback)
		return
	}
	d.DNSRouter.ExchangeAsync(ctx, message, options, func(response *mDNS.Msg, err error) {
		defer d.generation.release()
		callback(response, err)
	})
}

func (d *policyDNS) ClearCache() {
	if !d.acquire() {
		return
	}
	defer d.generation.release()
	d.DNSRouter.ClearCache()
}

func (d *policyDNS) LookupReverseMapping(ip netip.Addr) (string, bool) {
	if !d.acquire() {
		return (&accessDNS{policy: d.policy}).LookupReverseMapping(ip)
	}
	defer d.generation.release()
	return d.DNSRouter.LookupReverseMapping(ip)
}

func (d *policyDNS) ResetNetwork() {
	if !d.acquire() {
		// An old generation's network monitor must not reset the current
		// generation's DNS transports when its own policy has been retired.
		return
	}
	defer d.generation.release()
	d.DNSRouter.ResetNetwork()
}
