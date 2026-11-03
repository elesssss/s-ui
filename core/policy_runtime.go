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
	generation          *policyGeneration
	ctx                 context.Context
	scope               *adapter.Scope
	router              *route.Router
	dns                 *dns.Router
	transports          *manager.DNSTransportManager
	outbound            *manager.OutboundManager
	network             *route.NetworkManager
	connection          *route.ConnectionManager
	http                adapter.LifecycleService
	reference           *route.ReferenceManager
	closeOnce           sync.Once
	closed              atomic.Bool
	closeErr            error
	logger              log.ContextLogger
	components          map[adapter.Lifecycle]*adapter.Scope
	owned               []boxComponent
	networkInitializing bool
	routerInitializing  bool
	rulesStarting       bool
}

// Register ownership in construction order, before any subsequent operation
// can fail. Successful preparation replaces that order with dependency order.
func (p *policyRuntime) own(name string, lifecycle adapter.Lifecycle) *adapter.Scope {
	child := adapter.NewScope(p.ctx, p.logger)
	p.components[lifecycle] = child
	p.owned = append(p.owned, boxComponent{name, lifecycle})
	return child
}

func (p *policyRuntime) closeOwned() error {
	var err error
	for i := len(p.owned) - 1; i >= 0; i-- {
		component := p.owned[i]
		if closeErr := p.components[component.lifecycle].Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close %s: %w", component.name, closeErr))
		}
	}
	return err
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
		scope := p.scope
		if child := p.components[component.lifecycle]; child != nil {
			scope = child
		}
		if component.lifecycle == p.network && stage == adapter.StartStateInitialize {
			p.networkInitializing = true
		}
		if component.lifecycle == p.router {
			if stage == adapter.StartStateInitialize {
				p.routerInitializing = true
			}
			if stage == adapter.StartStatePostStart {
				p.rulesStarting = true
			}
		}
		if err := scope.Start(component.name, component.lifecycle, stage); err != nil {
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
	ctx = p.scope.Context()
	p.ctx = ctx
	p.components = make(map[adapter.Lifecycle]*adapter.Scope)
	p.scope.Add(p.closeOwned)
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
	p.outbound.Prepare(p.own("outbound", p.outbound))
	registry := service.FromContext[adapter.DNSTransportRegistry](ctx)
	p.transports = manager.NewDNSTransportManager(f.NewLogger("dns/transport"), registry, p.outbound, d.Final)
	p.transports.Prepare(p.own("dns-transport", p.transports))
	service.MustRegister[adapter.OutboundManager](ctx, p.outbound)
	service.MustRegister[adapter.DNSTransportManager](ctx, p.transports)
	p.dns, err = dns.NewRouter(ctx, f, d)
	if err != nil {
		return p, err
	}
	p.own("dns-router", p.dns)
	service.MustRegister[adapter.DNSRouter](ctx, &policyDNS{DNSRouter: p.dns, generation: p.generation, policy: b.policy})
	service.MustRegister[adapter.DNSRuleSetUpdateValidator](ctx, p.dns)
	// NetworkManager captures ConnectionManager in its constructor. Register
	// the candidate's owner first, so its notifications cannot reset old flows.
	p.connection = route.NewConnectionManager(ctx, f.NewLogger("connection"))
	p.own("connection", p.connection).Add(func() error { p.connection.CloseAll(); return nil })
	service.MustRegister[adapter.ConnectionManager](ctx, p.connection)
	p.network, err = route.NewNetworkManager(ctx, f.NewLogger("network"), r, d)
	if err != nil {
		return p, err
	}
	p.own("network", p.network).Add(func() error {
		if p.networkInitializing {
			return nil
		}
		return common.Close(p.network.InterfaceMonitor(), p.network.NetworkMonitor())
	})
	service.MustRegister[adapter.NetworkManager](ctx, p.network)
	h := httpclient.NewManager(ctx, f.NewLogger("httpclient"), options.HTTPClients, r.DefaultHTTPClient)
	p.http = h
	p.own(h.Name(), h)
	service.MustRegister[adapter.HTTPClientManager](ctx, h)
	p.router = route.NewRouter(ctx, f, r, d)
	p.own("router", p.router).Add(func() error {
		var cleanupErr error
		if !p.rulesStarting {
			for _, rule := range p.router.Rules() {
				cleanupErr = errors.Join(cleanupErr, rule.Close())
			}
		}
		if !p.routerInitializing {
			closed := make(map[string]bool)
			for _, options := range r.RuleSet {
				for _, tag := range options.Tag {
					if closed[tag] {
						continue
					}
					if ruleSet, found := p.router.RuleSet(tag); found {
						closed[tag] = true
						cleanupErr = errors.Join(cleanupErr, ruleSet.Close())
					}
				}
			}
		}
		return cleanupErr
	})
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
	p.own(p.reference.Name(), p.reference)
	// Stop rule downloads and probes before their DNS/network dependencies.
	p.owned = []boxComponent{
		{"network", p.network}, {"dns-transport", p.transports}, {"dns-router", p.dns},
		{"connection", p.connection}, {"outbound", p.outbound}, {h.Name(), h},
		{"router", p.router}, {p.reference.Name(), p.reference},
	}
	if err = p.startComponents(adapter.StartStateInitialize,
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
