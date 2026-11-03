package core

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/alireza0/s-ui/core/manager"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/inbound"
	"github.com/sagernet/sing-box/log"
)

type toggleTestInbound struct {
	inbound.Adapter
	closed   bool
	closeErr error
}

func (i *toggleTestInbound) Start(stage adapter.StartStage, scope *adapter.Scope) error {
	if stage == adapter.StartStateInitialize {
		scope.Add(func() error {
			i.closed = true
			return i.closeErr
		})
	}
	return nil
}

func TestRemoveInboundClosesOnlyItsConnections(t *testing.T) {
	for _, shutdown := range []string{"normal", "missing", "close-error"} {
		t.Run(shutdown, func(t *testing.T) {
			registry := inbound.NewRegistry()
			workers := make(map[string]*toggleTestInbound)
			inbound.Register[struct{}](registry, "test-toggle", func(_ context.Context, _ adapter.Router, _ log.ContextLogger, tag string, _ struct{}) (adapter.Inbound, error) {
				worker := &toggleTestInbound{Adapter: inbound.NewAdapter("test-toggle", tag)}
				workers[tag] = worker
				return worker, nil
			})
			logger := log.NewNOPFactory().Logger()
			inboundManager := manager.NewInboundManager(logger, registry, nil)
			for _, tag := range []string{"vpn", "other"} {
				if shutdown == "missing" && tag == "vpn" {
					continue
				}
				if err := inboundManager.Create(t.Context(), nil, logger, tag, "test-toggle", nil); err != nil {
					t.Fatal(err)
				}
			}
			scope := adapter.NewScope(t.Context(), logger)
			t.Cleanup(func() { _ = scope.Close() })
			if err := inboundManager.Start(adapter.StartStateInitialize, scope); err != nil {
				t.Fatal(err)
			}
			if shutdown == "close-error" {
				workers["vpn"].closeErr = errors.New("test scope close failed")
			}
			tracker := NewSessionTracker()
			vpnClient, vpnConn := trackConn(t, tracker, "vpn", "alice")
			defer vpnClient.Close()
			defer vpnConn.Close()
			otherClient, otherConn := trackConn(t, tracker, "other", "bob")
			defer otherClient.Close()
			defer otherConn.Close()
			c := &Core{isRunning: true, instance: &Box{inbound: inboundManager, sessionTracker: tracker}}
			err := c.RemoveInbound("vpn")
			if shutdown == "normal" && err != nil || shutdown == "missing" && !errors.Is(err, os.ErrInvalid) || shutdown == "close-error" && err == nil {
				t.Fatalf("unexpected removal result: %v", err)
			}
			if shutdown != "missing" && !workers["vpn"].closed {
				t.Fatal("removal left the listener scope open")
			}
			if workers["other"].closed || len(inboundManager.Inbounds()) != 1 {
				t.Fatal("removal closed the other inbound")
			}
			if _, err := vpnConn.Read(make([]byte, 1)); err == nil {
				t.Fatal("removed inbound connection remains open")
			}
			if sessions := tracker.Sessions(); len(sessions) != 1 || sessions[0].Inbound != "other" {
				t.Fatalf("removal affected the wrong sessions: %+v", sessions)
			}
		})
	}
}
