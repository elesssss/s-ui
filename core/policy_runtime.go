package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/alireza0/s-ui/core/manager"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/httpclient"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/dns"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/direct"
	"github.com/sagernet/sing-box/route"
	R "github.com/sagernet/sing-box/route/rule"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/service"
)

// Each policy owns a separate scope. Closing a retired policy must not cancel
// the box scope that owns access listeners, IKE/PPP and their gVisor devices.
type policyRuntime struct {
	generation *policyGeneration
	ctx        context.Context
	scope      *adapter.Scope
	router     *route.Router
	dns        *dns.Router
	transports *dns.TransportManager
	outbound   *manager.OutboundManager
	network    *route.NetworkManager
	connection *route.ConnectionManager
	http       adapter.LifecycleService
	reference  *route.ReferenceManager
	closeOnce  sync.Once
	closed     atomic.Bool
	closeErr   error
	logger     log.ContextLogger
}

func (p *policyRuntime) close() error {
	p.closeOnce.Do(func() {
		if p.scope != nil {
			p.closeErr = p.scope.Close()
		}
		if p.closeErr != nil && p.logger != nil {
			p.logger.Error("retire policy: ", p.closeErr)
		}
		p.closed.Store(true)
	})
	return p.closeErr
}

func (p *policyRuntime) startComponents(stage adapter.StartStage, components ...boxComponent) error {
	for _, component := range components {
		if err := p.scope.Start(component.name, component.lifecycle, stage); err != nil {
			return err
		}
	}
	return nil
}

func (b *Box) preparePolicy(options option.Options) (p *policyRuntime, err error) {
	// ExtendContext copies the registry; candidate dependencies never replace
	// those of the policy that is still serving traffic.
	ctx := service.ExtendContext(b.ctx)
	p = &policyRuntime{ctx: ctx, logger: b.logger, scope: adapter.NewScope(ctx, b.logger)}
	p.generation = &policyGeneration{runtime: p}
	defer func() {
		if err != nil {
			p.close()
		}
	}()
	r := common.PtrValueOrDefault(options.Route)
	d := common.PtrValueOrDefault(options.DNS)
	f := b.logFactory
	p.outbound = manager.NewOutboundManager(f.NewLogger("outbound"), service.FromContext[adapter.OutboundRegistry](ctx), b.endpoint, r.Final)
	registry := service.FromContext[adapter.DNSTransportRegistry](ctx)
	p.transports = dns.NewTransportManager(registry, p.outbound, d.Final)
	service.MustRegister[adapter.OutboundManager](ctx, p.outbound)
	service.MustRegister[adapter.DNSTransportManager](ctx, p.transports)
	p.dns, err = dns.NewRouter(ctx, f, d)
	if err != nil {
		return p, err
	}
	service.MustRegister[adapter.DNSRouter](ctx, &policyDNS{DNSRouter: p.dns, generation: p.generation, policy: b.policy})
	service.MustRegister[adapter.DNSRuleSetUpdateValidator](ctx, p.dns)
	// NetworkManager captures ConnectionManager in its constructor. Register
	// the candidate's owner first, so its notifications cannot reset old flows.
	p.connection = route.NewConnectionManager(ctx, f.NewLogger("connection"))
	service.MustRegister[adapter.ConnectionManager](ctx, p.connection)
	p.network, err = route.NewNetworkManager(ctx, f.NewLogger("network"), r, d)
	if err != nil {
		return p, err
	}
	service.MustRegister[adapter.NetworkManager](ctx, p.network)
	h := httpclient.NewManager(ctx, f.NewLogger("httpclient"), options.HTTPClients, r.DefaultHTTPClient)
	p.http = h
	service.MustRegister[adapter.HTTPClientManager](ctx, h)
	p.router = route.NewRouter(ctx, f, r, d)
	service.MustRegister[adapter.Router](ctx, p.router)
	if err = p.router.Initialize(r.Rules, r.RuleSet); err != nil {
		return p, err
	}
	if err = p.dns.Initialize(d.Rules); err != nil {
		return p, err
	}
	for i, o := range d.Servers {
		tag := o.Tag
		if tag == "" {
			tag = fmt.Sprint(i)
		}
		if err = p.transports.Create(ctx, f.NewLogger("dns/"+tag), tag, o.Type, o.Options); err != nil {
			return p, err
		}
	}
	for i, o := range options.Outbounds {
		tag := o.Tag
		if tag == "" {
			tag = fmt.Sprint(i)
		}
		outCtx := adapter.WithContext(ctx, &adapter.InboundContext{Outbound: tag})
		if err = p.outbound.Create(outCtx, p.router, f.NewLogger("outbound/"+tag), tag, o.Type, o.Options); err != nil {
			return p, err
		}
	}
	p.outbound.Initialize(func() (adapter.Outbound, error) {
		return direct.NewOutbound(ctx, p.router, f.NewLogger("outbound/direct"), "direct", option.DirectOutboundOptions{})
	})
	p.transports.Initialize(func() (adapter.DNSTransport, error) {
		return registry.CreateDNSTransport(ctx, f.NewLogger("dns/local"), "local", C.DNSTypeLocal, &option.LocalDNSServerOptions{})
	})
	h.Initialize(func() (*httpclient.ManagedTransport, error) {
		return httpclient.NewTransport(ctx, f.NewLogger("httpclient"), "", option.HTTPClientOptions{DefaultOutbound: true})
	})
	p.router.AppendTracker(b.sessionTracker)
	p.reference = route.NewReferenceManager(ctx, f.NewLogger("reference"), options)
	// First registration controls scope cleanup order. Stop rules, HTTP clients
	// and outbounds before disposing the DNS/network services they depend on.
	if err = p.startComponents(adapter.StartStateInitialize,
		boxComponent{"network", p.network},
		boxComponent{"dns-transport", p.transports},
		boxComponent{"dns-router", p.dns},
		boxComponent{"connection", p.connection},
		boxComponent{"outbound", p.outbound},
		boxComponent{h.Name(), h},
		boxComponent{"router", p.router},
		boxComponent{p.reference.Name(), p.reference},
	); err != nil {
		return p, err
	}
	for _, stage := range []adapter.StartStage{adapter.StartStateStart, adapter.StartStatePostStart, adapter.StartStateStarted} {
		if err = p.startComponents(stage,
			boxComponent{"outbound", p.outbound},
			boxComponent{"dns-transport", p.transports},
			boxComponent{"network", p.network},
			boxComponent{"connection", p.connection},
			boxComponent{h.Name(), h},
			boxComponent{"router", p.router},
			boxComponent{"dns-router", p.dns},
			boxComponent{p.reference.Name(), p.reference},
		); err != nil {
			return p, err
		}
	}
	for _, rule := range p.router.Rules() {
		var target string
		switch action := rule.Action().(type) {
		case *R.RuleActionRoute:
			target = action.Outbound
		case *R.RuleActionBypass:
			target = action.Outbound
		}
		if target != "" {
			if _, found := p.outbound.Outbound(target); !found {
				return p, fmt.Errorf("route references missing outbound %q", target)
			}
		}
	}
	if p.outbound.Default() == nil {
		return p, errors.New("missing default outbound")
	}
	return p, nil
}
