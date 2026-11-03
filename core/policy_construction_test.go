package core

import (
	"context"
	"strings"
	"testing"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/service"
)

type constructionOutbound struct {
	adapter.Outbound
	initialized int
	closed      int
	ctx         context.Context
}

func (*constructionOutbound) Type() string           { return "construction-probe" }
func (*constructionOutbound) Tag() string            { return "one" }
func (*constructionOutbound) Dependencies() []string { return nil }
func (o *constructionOutbound) Start(stage adapter.StartStage, scope *adapter.Scope) error {
	if stage == adapter.StartStateInitialize {
		o.initialized++
		scope.Add(func() error { o.closed++; return nil })
	}
	return nil
}

// Fail before native network startup: resources created before a later
// constructor/default-selection error must close without touching access.
func TestPolicyConstructionFailureClosesCandidateOnly(t *testing.T) {
	for _, raw := range []string{
		`{"outbounds":[{"type":"construction-probe","tag":"one"},{"type":"direct","tag":"bad","detour":"invalid"}]}`,
		`{"outbounds":[{"type":"construction-probe","tag":"one"}],"dns":{"final":"missing","servers":[{"type":"hosts","tag":"hosts"}]}}`,
	} {
		t.Run(raw, func(t *testing.T) {
			c := NewCore()
			probe := &constructionOutbound{}
			registry := service.FromContext[adapter.OutboundRegistry](c.ctx).(*outbound.Registry)
			outbound.Register[struct{}](registry, "construction-probe", func(ctx context.Context, _ adapter.Router, _ log.ContextLogger, _ string, _ struct{}) (adapter.Outbound, error) {
				probe.ctx = ctx
				return probe, nil
			})
			box, err := NewBox(Options{Context: c.ctx})
			if err != nil {
				t.Fatal(err)
			}
			defer box.Close()
			var options option.Options
			if err := options.UnmarshalJSONContext(c.ctx, []byte(raw)); err != nil {
				t.Fatal(err)
			}
			p, err := box.preparePolicy(options)
			if err == nil {
				p.close()
				t.Fatal("invalid candidate succeeded")
			}
			if strings.Contains(err.Error(), "netlink") {
				t.Fatalf("failure reached network startup: %v", err)
			}
			if !p.closed.Load() || probe.closed != 1 || probe.initialized != 1 {
				t.Fatalf("candidate cleanup: closed=%v resource=%d initialized=%d", p.closed.Load(), probe.closed, probe.initialized)
			}
			if p.ctx.Err() == nil || probe.ctx.Err() == nil {
				t.Fatal("candidate construction context survived cleanup")
			}
			if box.ctx.Err() != nil || box.scope.Context().Err() != nil || box.policy.current.runtime.closed.Load() {
				t.Fatal("candidate failure closed access or current policy")
			}
			if err := p.close(); err != nil {
				t.Fatal(err)
			}
			if probe.closed != 1 {
				t.Fatal("candidate resource closed twice")
			}
		})
	}
}

func TestPolicyAccessContextOutlivesRetiredCandidate(t *testing.T) {
	logger := log.NewNOPFactory().Logger()
	accessScope := adapter.NewScope(context.Background(), logger)
	defer accessScope.Close()
	p := &policyRuntime{scope: adapter.NewScope(accessScope.Context(), logger)}
	p.ctx = service.ContextWithDefaultRegistry(p.scope.Context())
	service.MustRegister[int](p.ctx, 1)
	p.generation = &policyGeneration{runtime: p}
	box := &Box{scope: accessScope, policy: &policyRouter{current: p.generation}}
	ctx, err := box.accessContext()
	if err != nil {
		t.Fatal(err)
	}
	service.MustRegister[int](ctx, 2)
	if service.FromContext[int](p.ctx) != 1 {
		t.Fatal("access context changed the policy registry")
	}
	p.generation.retire(policyImpact{})
	if p.ctx.Err() == nil {
		t.Fatal("retired candidate context remained live")
	}
	if ctx.Err() != nil {
		t.Fatal("retiring the policy cancelled access")
	}
	if err := accessScope.Close(); err != nil {
		t.Fatal(err)
	}
	if ctx.Err() == nil {
		t.Fatal("stopping the box did not cancel access")
	}
}
