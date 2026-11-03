package sub

import (
	"encoding/json"
	"testing"

	"github.com/alireza0/s-ui/database/model"
)

// The source skips empty out_json, rather than imposing an L2TP-specific
// filter. Explicit historical out_json still follows the generic path.
func TestL2TPSubscriptionUsesStoredOutbound(t *testing.T) {
	service := &JsonService{}
	for _, present := range []bool{false, true} {
		inbound := &model.Inbound{Type: "l2tp", Tag: "vpn", Addrs: json.RawMessage(`[]`)}
		if present {
			inbound.OutJson = json.RawMessage(`{"type":"l2tp","tag":"vpn"}`)
		}
		outbounds, tags, err := service.getOutbounds(json.RawMessage(`{}`), []*model.Inbound{inbound}, "")
		if err != nil {
			t.Fatal(err)
		}
		want := 0
		if present {
			want = 1
		}
		if len(*outbounds) != want || len(*tags) != want {
			t.Fatalf("stored=%v: outbounds=%v tags=%v", present, *outbounds, *tags)
		}
	}
}
