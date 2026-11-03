package model

import (
	"encoding/json"
	"testing"
)

func TestL2TPLegacyEnableIsIgnored(t *testing.T) {
	inbound := Inbound{Type: "l2tp", Tag: "vpn", Options: json.RawMessage(`{"enable":false,"psk":"test-only"}`)}
	config, err := inbound.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(config, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["enable"]; present {
		t.Fatal("legacy option reached the core")
	}
	full, err := inbound.MarshalFull()
	if err != nil {
		t.Fatal(err)
	}
	if _, present := (*full)["enable"]; present {
		t.Fatal("legacy option reached the form")
	}
	if err := inbound.UnmarshalJSON([]byte(`{"type":"l2tp","enable":false,"psk":"test-only"}`)); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(inbound.Options, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["enable"]; present {
		t.Fatal("legacy option was stored again")
	}
	if raw["psk"] != "test-only" {
		t.Fatal("protocol options were lost")
	}
	if err := inbound.UnmarshalJSON([]byte(`{"type":"vless","enable":false}`)); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(inbound.Options, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["enable"] != false {
		t.Fatal("another protocol lost its enable option")
	}
}
