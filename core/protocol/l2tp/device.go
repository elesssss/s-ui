//go:build linux && with_gvisor

package l2tp

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/sagernet/gvisor/pkg/buffer"
	"github.com/sagernet/gvisor/pkg/tcpip/header"
	"github.com/sagernet/gvisor/pkg/tcpip/link/channel"
	gs "github.com/sagernet/gvisor/pkg/tcpip/stack"
	tun "github.com/sagernet/sing-tun"
)

// The bounded link queue applies backpressure without touching host networking.
type packetDevice struct {
	ctx      context.Context
	cancel   context.CancelFunc
	endpoint *channel.Endpoint
	stack    tun.Stack
	once     sync.Once
}

func newPacketDevice(parent context.Context, mtu uint16) *packetDevice {
	ctx, cancel := context.WithCancel(parent)
	return &packetDevice{ctx: ctx, cancel: cancel, endpoint: channel.New(128, uint32(mtu), "")}
}
func (d *packetDevice) Name() (string, error)                { return "l2tp-memory", nil }
func (d *packetDevice) Start() error                         { return nil }
func (d *packetDevice) UpdateRouteOptions(tun.Options) error { return nil }
func (d *packetDevice) NewEndpoint() (gs.LinkEndpoint, gs.NICOptions, error) {
	return d.endpoint, gs.NICOptions{}, nil
}
func (d *packetDevice) Read(b []byte) (int, error) {
	pkt := d.endpoint.ReadContext(d.ctx)
	if pkt == nil {
		return 0, net.ErrClosed
	}
	defer pkt.DecRef()
	view := pkt.ToView()
	defer view.Release()
	if view.Size() > len(b) {
		return 0, io.ErrShortBuffer
	}
	return copy(b, view.AsSlice()), nil
}
func (d *packetDevice) Write(b []byte) (int, error) {
	if d.ctx.Err() != nil {
		return 0, net.ErrClosed
	}
	if len(b) < 20 || b[0]>>4 != 4 {
		return 0, io.ErrUnexpectedEOF
	}
	pkt := gs.NewPacketBuffer(gs.PacketBufferOptions{Payload: buffer.MakeWithData(append([]byte(nil), b...))})
	defer pkt.DecRef()
	d.endpoint.InjectInbound(header.IPv4ProtocolNumber, pkt)
	return len(b), nil
}
func (d *packetDevice) WritePacket(pkt *gs.PacketBuffer) (int, error) {
	var packets gs.PacketBufferList
	packets.PushBack(pkt)
	n, err := d.endpoint.WritePackets(packets)
	if err != nil {
		return n, net.ErrClosed
	}
	return pkt.Size(), nil
}
func (d *packetDevice) Close() error {
	d.once.Do(func() {
		d.cancel()
		d.endpoint.Close()
		if d.stack != nil {
			_ = d.stack.Close()
		}
	})
	return nil
}
