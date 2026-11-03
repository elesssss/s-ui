package service

import (
	"encoding/json"
	"testing"

	"github.com/alireza0/s-ui/core/protocol/l2tp"
	"github.com/alireza0/s-ui/database/model"
)

// Adding the protocol to an existing database must select only enabled users
// with L2TP credentials, while preserving the existing client-id filtering.
func TestL2TPUsersExcludeDisabledAndLegacyClients(t *testing.T) {
	db := initUsersTestDB(t)
	for _, client := range []model.Client{
		{Name: "vpn-user", Enable: true, Config: json.RawMessage(`{"l2tp":{"name":"vpn-user","password":"test-only"}}`)},
		{Name: "disabled-vpn", Enable: false, Config: json.RawMessage(`{"l2tp":{"name":"disabled-vpn","password":"test-only"}}`)},
	} {
		if err := db.Create(&client).Error; err != nil {
			t.Fatal(err)
		}
	}
	s := &InboundService{}
	out, err := s.initUsers(db, json.RawMessage(`{"type":"l2tp"}`), "1,2,3,4", "l2tp")
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		Users []l2tp.User `json:"users"`
	}
	if err := json.Unmarshal(out, &config); err != nil {
		t.Fatal(err)
	}
	if len(config.Users) != 1 || config.Users[0].Name != "vpn-user" || config.Users[0].Password != "test-only" {
		t.Fatalf("wrong enabled L2TP users: %+v", config.Users)
	}
	out, err = s.initUsers(db, json.RawMessage(`{"type":"l2tp"}`), "1,2", "l2tp")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(out, &config); err != nil {
		t.Fatal(err)
	}
	if len(config.Users) != 0 {
		t.Fatalf("legacy users inherited credentials: %+v", config.Users)
	}
}
