package service

import (
	"encoding/json"
	"testing"

	"github.com/alireza0/s-ui/core/protocol/l2tp"
	"github.com/alireza0/s-ui/database/model"
)

// Match the source selection: only enable and requested IDs filter users.
func TestL2TPUsersExcludeDisabledClients(t *testing.T) {
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
	if err := db.Model(model.Client{}).Where("name = ?", "disabled-vpn").Update("enable", false).Error; err != nil {
		t.Fatal(err)
	}
	out, err := s.initUsers(db, json.RawMessage(`{"type":"l2tp"}`), "3,4", "l2tp")
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
	// Legacy clients are no longer silently removed from the selection. Like
	// the source, their missing protocol config makes configuration fail.
	if _, err = s.initUsers(db, json.RawMessage(`{"type":"l2tp"}`), "1,2", "l2tp"); err == nil {
		t.Fatal("legacy clients without L2TP credentials were silently ignored")
	}
}
