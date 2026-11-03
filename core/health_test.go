package core

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/alireza0/s-ui/core/manager"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/inbound"
	"github.com/sagernet/sing-box/log"
)

type healthInbound struct {
	inbound.Adapter
	err error
}

func (i *healthInbound) Start(adapter.StartStage, *adapter.Scope) error { return nil }

func (i *healthInbound) HealthError() error { return i.err }

func TestCoreHealthIncludesInboundWorkers(t *testing.T) {
	registry := inbound.NewRegistry()
	worker := &healthInbound{Adapter: inbound.NewAdapter("test-health", "vpn")}
	inbound.Register[struct{}](registry, "test-health", func(context.Context, adapter.Router, log.ContextLogger, string, struct{}) (adapter.Inbound, error) {
		return worker, nil
	})
	managerInstance := manager.NewInboundManager(log.NewNOPFactory().Logger(), registry, nil)
	if err := managerInstance.Create(context.Background(), nil, log.NewNOPFactory().Logger(), "vpn", "test-health", nil); err != nil {
		t.Fatal(err)
	}
	c := &Core{isRunning: true, instance: &Box{inbound: managerInstance}}
	if !c.IsRunning() || c.HealthError() != nil {
		t.Fatal("healthy worker marked failed")
	}
	cause := errors.New("listener read failed")
	worker.err = cause
	if c.IsRunning() || !errors.Is(c.HealthError(), cause) || !strings.Contains(c.HealthError().Error(), "vpn") {
		t.Fatal("worker failure absent from core health")
	}
	if c.GetInstance() == nil {
		t.Fatal("unhealthy box must remain available for cleanup")
	}
	worker.err = nil
	if !c.IsRunning() {
		t.Fatal("replacement healthy worker remains failed")
	}
}
