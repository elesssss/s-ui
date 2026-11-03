package l2tp

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/inbound"
	"github.com/sagernet/sing-box/log"
)

const Type = "l2tp"

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Options struct {
	Listen     string `json:"listen,omitempty"`
	ListenPort uint16 `json:"listen_port,omitempty"`
	PublicIP   string `json:"public_ip"`
	PSK        string `json:"psk"`
	Pool       string `json:"pool,omitempty"`
	Users      []User `json:"users,omitempty"`
}

func (o *Options) validate() error {
	if o.Listen == "" {
		o.Listen = "0.0.0.0"
	}
	if a, err := netip.ParseAddr(o.Listen); err != nil || !a.Is4() {
		return fmt.Errorf("l2tp: listen must be an IPv4 address")
	}
	if a, err := netip.ParseAddr(o.PublicIP); err != nil || !a.Is4() || a.IsUnspecified() || a.IsMulticast() {
		return fmt.Errorf("l2tp: public_ip must be a concrete IPv4 address")
	}
	if o.ListenPort == 0 {
		o.ListenPort = 500
	}
	if o.ListenPort == 4500 {
		return fmt.Errorf("l2tp: IKE port conflicts with NAT-T port 4500")
	}
	if o.PSK == "" {
		return fmt.Errorf("l2tp: PSK is required")
	}
	if o.Pool == "" {
		o.Pool = "10.20.0.0/24"
	}
	p, err := netip.ParsePrefix(o.Pool)
	if err != nil || !p.Addr().Is4() || p.Bits() < 16 || p.Bits() > 29 || p != p.Masked() || !p.Addr().IsPrivate() {
		return fmt.Errorf("l2tp: pool must be a canonical private IPv4 /16 through /29 subnet")
	}
	names := make(map[string]bool)
	for _, u := range o.Users {
		if u.Name == "" || u.Password == "" || names[u.Name] {
			return fmt.Errorf("l2tp: empty or duplicate user credentials")
		}
		names[u.Name] = true
	}
	return nil
}

func RegisterInbound(registry *inbound.Registry) {
	inbound.Register[Options](registry, Type, func(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options Options) (adapter.Inbound, error) {
		if err := options.validate(); err != nil {
			return nil, err
		}
		return newInbound(ctx, router, logger, tag, options)
	})
}
