package service

import (
	"encoding/json"
	"testing"

	"github.com/alireza0/s-ui/database/model"
)

func TestResetAllClientsTraffic(t *testing.T) {
	db := settingTestDB(t)
	clients := []model.Client{
		{Name: "depleted", Enable: true, Up: 10, Down: 10, Inbounds: json.RawMessage(`[1,2]`)},
		{Name: "active", Enable: true, Up: 5, Inbounds: json.RawMessage(`[3]`)},
		{Name: "idle", Enable: true, Inbounds: json.RawMessage(`[4]`)},
	}
	for i := range clients {
		clients[i].Config = json.RawMessage(`{}`)
		clients[i].Links = json.RawMessage(`[]`)
		if err := db.Create(&clients[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	// Created enabled, then disabled: a false bool is GORM's zero value and
	// would be replaced by the column default on create.
	if err := db.Model(model.Client{}).Where("name = ?", "depleted").Update("enable", false).Error; err != nil {
		t.Fatal(err)
	}

	err := (&ClientService{}).ResetAllClientsTraffic()
	if err != nil {
		t.Fatal(err)
	}

	var stored []model.Client
	if err := db.Order("id").Find(&stored).Error; err != nil {
		t.Fatal(err)
	}
	for _, c := range stored {
		if !c.Enable || c.Up+c.Down != 0 {
			t.Errorf("%s: expected enabled with zero traffic, got enable=%v up=%d down=%d", c.Name, c.Enable, c.Up, c.Down)
		}
	}
	if stored[0].TotalUp != 10 || stored[0].TotalDown != 10 {
		t.Errorf("depleted: traffic must accumulate into the totals, got %d/%d", stored[0].TotalUp, stored[0].TotalDown)
	}
}
