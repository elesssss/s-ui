package core

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	mDNS "github.com/miekg/dns"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing/service"
)

// Access listeners may cache their DNSRouter at construction. Keep that handle
// stable, but resolve named transports in the selected generation, never in the
// retired generation from which the listener obtained its options.
type accessDNS struct {
	adapter.DNSRouter
	policy *policyRouter
}

func (d *accessDNS) options(p *policyRuntime, o adapter.DNSQueryOptions) (adapter.DNSQueryOptions, error) {
	if o.Transport != nil {
		tag := o.Transport.Tag()
		var found bool
		o.Transport, found = p.transports.Transport(tag)
		if !found {
			return o, fmt.Errorf("DNS transport %q is no longer configured", tag)
		}
	}
	return o, nil
}
func (d *accessDNS) Exchange(ctx context.Context, m *mDNS.Msg, o adapter.DNSQueryOptions) (*mDNS.Msg, error) {
	g := d.policy.acquire()
	if g == nil {
		return nil, net.ErrClosed
	}
	defer g.release()
	o, err := d.options(g.runtime, o)
	if err != nil {
		return nil, err
	}
	return g.runtime.dns.Exchange(ctx, m, o)
}
func (d *accessDNS) ExchangeAsync(ctx context.Context, m *mDNS.Msg, o adapter.DNSQueryOptions, callback func(*mDNS.Msg, error)) {
	g := d.policy.acquire()
	if g == nil {
		callback(nil, net.ErrClosed)
		return
	}
	o, err := d.options(g.runtime, o)
	if err != nil {
		g.release()
		callback(nil, err)
		return
	}
	g.runtime.dns.ExchangeAsync(ctx, m, o, func(m *mDNS.Msg, err error) { defer g.release(); callback(m, err) })
}
func (d *accessDNS) Lookup(ctx context.Context, domain string, o adapter.DNSQueryOptions) ([]netip.Addr, error) {
	g := d.policy.acquire()
	if g == nil {
		return nil, net.ErrClosed
	}
	defer g.release()
	o, err := d.options(g.runtime, o)
	if err != nil {
		return nil, err
	}
	return g.runtime.dns.Lookup(ctx, domain, o)
}
func (d *accessDNS) ClearCache() {
	g := d.policy.acquire()
	if g == nil {
		return
	}
	defer g.release()
	g.runtime.dns.ClearCache()
}
func (d *accessDNS) LookupReverseMapping(ip netip.Addr) (string, bool) {
	g := d.policy.acquire()
	if g == nil {
		return "", false
	}
	defer g.release()
	return g.runtime.dns.LookupReverseMapping(ip)
}
func (d *accessDNS) ResetNetwork() {
	g := d.policy.acquire()
	if g == nil {
		return
	}
	defer g.release()
	g.runtime.dns.ResetNetwork()
}

func (b *Box) accessContext() (context.Context, error) {
	g := b.policy.acquire()
	if g == nil {
		return nil, net.ErrClosed
	}
	defer g.release()
	ctx := service.ExtendContext(g.runtime.ctx)
	service.MustRegister[adapter.Router](ctx, b.policy)
	service.MustRegister[adapter.NetworkManager](ctx, b.network)
	service.MustRegister[adapter.DNSRouter](ctx, &accessDNS{DNSRouter: b.dnsRouter, policy: b.policy})
	return ctx, nil
}
