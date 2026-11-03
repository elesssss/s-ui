package model

import (
	"encoding/json"
	"testing"
)

func TestInboundEnablePersistsOnlyInPanelConfig(t *testing.T) {
	for _, protocol := range []string{"l2tp", "vless", "direct", "tun", "cloudflared"} {
		t.Run(protocol, func(t *testing.T) {
			var inbound Inbound
			input := []byte(`{"type":"` + protocol + `","tag":"test","enable":false,"listen_port":500}`)
			if err := inbound.UnmarshalJSON(input); err != nil {
				t.Fatal(err)
			}
			enabled, err := inbound.IsEnabled()
			if err != nil || enabled {
				t.Fatalf("disabled flag lost: %v %v", enabled, err)
			}
			config, err := inbound.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			var raw map[string]any
			if err := json.Unmarshal(config, &raw); err != nil {
				t.Fatal(err)
			}
			if _, present := raw["enable"]; present {
				t.Fatal("panel enable field reached sing-box")
			}
			full, err := inbound.MarshalFull()
			if err != nil || (*full)["enable"] != false {
				t.Fatalf("form lost disabled flag: %v %v", full, err)
			}
			if err := inbound.SetEnabled(true); err != nil {
				t.Fatal(err)
			}
			full, err = inbound.MarshalFull()
			if err != nil || (*full)["enable"] != true || (*full)["listen_port"] != float64(500) {
				t.Fatalf("toggle changed protocol options: %v %v", full, err)
			}
		})
	}
}

func TestInboundEnableDefaultsAndValidation(t *testing.T) {
	for _, options := range []string{"", "null", `{}`, `{"enable":null}`, `{"enable":true}`, `{"enable":false}`} {
		t.Run(options, func(t *testing.T) {
			inbound := Inbound{Options: json.RawMessage(options)}
			want := options != `{"enable":false}`
			enabled, err := inbound.IsEnabled()
			if err != nil || enabled != want {
				t.Fatalf("enabled = %v, %v; want %v", enabled, err, want)
			}
			full, err := inbound.MarshalFull()
			if err != nil || (*full)["enable"] != want {
				t.Fatalf("form = %v, %v; want enable %v", full, err, want)
			}
		})
	}
	for _, options := range []string{`{"enable":"false"}`, `{"enable":0}`, `{"enable":[]}`, `[]`, `{`} {
		inbound := Inbound{Options: json.RawMessage(options)}
		if _, err := inbound.IsEnabled(); err == nil {
			t.Fatalf("invalid enable metadata accepted: %s", options)
		}
	}
	var inbound Inbound
	if err := inbound.UnmarshalJSON([]byte(`{"type":"l2tp","enable":"false"}`)); err == nil {
		t.Fatal("invalid flag accepted on save")
	}
}
