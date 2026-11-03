package sub

import (
	"encoding/json"
	"testing"

	"github.com/alireza0/s-ui/database/model"
)

// Skip L2TP even if an API caller supplied out_json rather than letting the
// inbound service clear it. It is not a sing-box/Clash client protocol.
func TestL2TPExcludedFromSubscriptionOutbounds(t *testing.T) {
	service := &JsonService{}
	outbounds, tags, err := service.getOutbounds(json.RawMessage(`{}`), []*model.Inbound{{
		Type: "l2tp", Tag: "vpn", OutJson: json.RawMessage(`{"type":"l2tp","psk":"test-only"}`),
	}}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(*outbounds) != 0 || len(*tags) != 0 {
		t.Fatalf("L2TP leaked into subscriptions: %v %v", *outbounds, *tags)
	}
}
