//go:build !linux || !with_gvisor

package l2tp

import (
	"context"
	"fmt"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/log"
)

func newInbound(context.Context, adapter.Router, log.ContextLogger, string, Options) (adapter.Inbound, error) {
	return nil, fmt.Errorf("l2tp: requires Linux and the with_gvisor build profile (use build.sh)")
}
