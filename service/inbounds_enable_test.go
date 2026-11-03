package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"

	"gorm.io/gorm"
)

// A listener registry with no sockets. It exercises the real database and
// runtime coordination while keeping VPN, firewall and host networking idle.
type inboundTestRuntime struct {
	configs map[string][]byte
	calls   []string
	addErr  error
}

func (r *inboundTestRuntime) AddInbound(config []byte) error {
	var inbound struct {
		Tag string `json:"tag"`
	}
	if err := json.Unmarshal(config, &inbound); err != nil {
		return err
	}
	r.calls = append(r.calls, "add "+inbound.Tag)
	if r.addErr != nil {
		err := r.addErr
		r.addErr = nil
		return err
	}
	if _, exists := r.configs[inbound.Tag]; exists {
		return errors.New("duplicate inbound")
	}
	r.configs[inbound.Tag] = append([]byte(nil), config...)
	return nil
}

func (r *inboundTestRuntime) RemoveInbound(tag string) error {
	r.calls = append(r.calls, "remove "+tag)
	if _, exists := r.configs[tag]; !exists {
		return os.ErrInvalid
	}
	delete(r.configs, tag)
	return nil
}

func inboundEnableFixture(t *testing.T, enabled bool) (*InboundService, *gorm.DB, model.Inbound, *inboundTestRuntime) {
	t.Helper()
	lifecycleService(t)
	db := database.GetDB()
	inbound := model.Inbound{
		Type: "l2tp", Tag: "vpn",
		Options: json.RawMessage(`{"listen":"0.0.0.0","listen_port":500,"public_ip":"192.0.2.1","pool":"10.20.0.0/24","psk":"test-only"}`),
		Addrs:   json.RawMessage(`[]`), OutJson: json.RawMessage(`null`),
	}
	if err := inbound.SetEnabled(enabled); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&inbound).Error; err != nil {
		t.Fatal(err)
	}
	client := model.Client{
		Name: "vpn-user", Enable: true,
		Inbounds: json.RawMessage(fmt.Sprintf("[%d]", inbound.Id)),
		Config:   json.RawMessage(`{"l2tp":{"name":"vpn-user","password":"test-only"}}`),
		Links:    json.RawMessage(`[{"remark":"keep-me","type":"remote","uri":"https://example.test"}]`),
	}
	if err := db.Create(&client).Error; err != nil {
		t.Fatal(err)
	}
	s := &InboundService{}
	runtime := &inboundTestRuntime{configs: make(map[string][]byte)}
	if enabled {
		config, err := s.runtimeConfig(db, &inbound)
		if err != nil {
			t.Fatal(err)
		}
		runtime.configs[inbound.Tag] = config
	}
	return s, db, inbound, runtime
}

func inboundToggleData(id uint, enable bool) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"id":%d,"enable":%t,"tag":"stale-tag","psk":"stale-secret"}`, id, enable))
}

func storedInboundEnabled(t *testing.T, db *gorm.DB, id uint) bool {
	t.Helper()
	var inbound model.Inbound
	if err := db.First(&inbound, id).Error; err != nil {
		t.Fatal(err)
	}
	enabled, err := inbound.IsEnabled()
	if err != nil {
		t.Fatal(err)
	}
	return enabled
}

func TestInboundTogglePreservesConfigAndUserBindings(t *testing.T) {
	s, db, inbound, runtime := inboundEnableFixture(t, true)
	var clientBefore, clientAfter model.Client
	if err := db.First(&clientBefore).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := s.saveAndApply(db, runtime, "toggle", inboundToggleData(inbound.Id, false), "", "test", ""); err != nil {
		t.Fatal(err)
	}
	if storedInboundEnabled(t, db, inbound.Id) || len(runtime.configs) != 0 {
		t.Fatal("disabled inbound was kept running or not persisted")
	}
	var stored model.Inbound
	if err := db.First(&stored, inbound.Id).Error; err != nil {
		t.Fatal(err)
	}
	if err := stored.SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stored, inbound) {
		t.Fatalf("switch overwrote inbound config: %+v", stored)
	}
	if err := db.First(&clientAfter).Error; err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(clientBefore, clientAfter) {
		t.Fatal("switch changed client credentials, links or bindings")
	}
	list, err := s.GetAll()
	if err != nil || len(*list) != 1 || (*list)[0]["enable"] != false {
		t.Fatalf("list lost disabled state: %v %v", list, err)
	}
	// A repeated switch is harmless and does not restart any listener.
	calls := len(runtime.calls)
	if _, err := s.saveAndApply(db, runtime, "toggle", inboundToggleData(inbound.Id, false), "", "test", ""); err != nil || len(runtime.calls) != calls {
		t.Fatalf("repeated disable changed runtime: %v", err)
	}
	if err := db.Model(&clientAfter).Update("config", json.RawMessage(`{"l2tp":{"name":"vpn-user","password":"updated-test-only"}}`)).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := s.saveAndApply(db, runtime, "toggle", inboundToggleData(inbound.Id, true), "", "test", ""); err != nil {
		t.Fatal(err)
	}
	var config struct {
		Users []struct{ Name, Password string } `json:"users"`
	}
	if err := json.Unmarshal(runtime.configs["vpn"], &config); err != nil {
		t.Fatal(err)
	}
	if !storedInboundEnabled(t, db, inbound.Id) || len(config.Users) != 1 || config.Users[0].Password != "updated-test-only" {
		t.Fatal("reenabling did not use the current bound user credentials")
	}
}

func TestDisabledInboundsExcludedFromCoreAndUpdateSelection(t *testing.T) {
	s, db, disabled, _ := inboundEnableFixture(t, false)
	// Missing L2TP credentials would fail addUsers if this inbound were loaded.
	if err := db.Model(&model.Client{}).Where("name = ?", "vpn-user").Update("config", json.RawMessage(`{}`)).Error; err != nil {
		t.Fatal(err)
	}
	for _, inbound := range []model.Inbound{
		{Type: "direct", Tag: "legacy", Options: json.RawMessage(`{}`)},
		{Type: "direct", Tag: "null-enable", Options: json.RawMessage(`{"enable":null}`)},
	} {
		if err := db.Create(&inbound).Error; err != nil {
			t.Fatal(err)
		}
	}
	config, err := s.GetAllConfig(db)
	if err != nil || len(config) != 2 {
		t.Fatalf("disabled inbound affected core startup: %v %v", config, err)
	}
	for _, raw := range config {
		var fields map[string]any
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		if fields["tag"] == "vpn" || fields["enable"] != nil {
			t.Fatal("disabled inbound or panel metadata reached core config")
		}
	}
	// Both user updates and TLS restarts use enabledInbounds with the IDs.
	for _, ids := range [][]uint{{disabled.Id}, {}} {
		active, err := s.enabledInbounds(db, ids)
		if err != nil || len(active) != 0 {
			t.Fatalf("disabled inbound selected for user/TLS updates: %v %v", active, err)
		}
	}
}

func TestInboundToggleRejectsInvalidRequests(t *testing.T) {
	s, db, inbound, runtime := inboundEnableFixture(t, true)
	for _, data := range []string{`{}`, `{"id":0,"enable":false}`, `{"id":1}`, `{"id":1,"enable":null}`, `{"id":1,"enable":"false"}`, `{"id":99999,"enable":false}`} {
		if _, err := s.saveAndApply(db, runtime, "toggle", []byte(data), "", "test", ""); err == nil {
			t.Fatalf("invalid toggle accepted: %s", data)
		}
	}
	if !storedInboundEnabled(t, db, inbound.Id) || len(runtime.calls) != 0 {
		t.Fatal("invalid request changed inbound state")
	}
}

type inboundCommitFailurePool struct{ *sql.DB }

func (p inboundCommitFailurePool) BeginTx(ctx context.Context, options *sql.TxOptions) (gorm.ConnPool, error) {
	tx, err := p.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return &inboundCommitFailureTx{tx}, nil
}

type inboundCommitFailureTx struct{ *sql.Tx }

func (tx inboundCommitFailureTx) Commit() error {
	_ = tx.Tx.Rollback()
	return errors.New("test inbound commit failed")
}

func TestInboundToggleFailurePreservesDatabaseAndRuntime(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		for _, failure := range []string{"audit", "commit", "activation"} {
			if failure == "activation" && enabled {
				continue // Disabling creates no replacement listener.
			}
			t.Run(fmt.Sprintf("enabled=%t/%s", enabled, failure), func(t *testing.T) {
				s, db, inbound, runtime := inboundEnableFixture(t, enabled)
				switch failure {
				case "audit":
					const callback = "test:fail_inbound_audit"
					if err := db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
						if tx.Statement.Schema != nil && tx.Statement.Schema.Table == "changes" {
							tx.AddError(errors.New("test audit write failed"))
						}
					}); err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { db.Callback().Create().Remove(callback) })
				case "commit":
					sqlDB, err := db.DB()
					if err != nil {
						t.Fatal(err)
					}
					pool, statementPool := db.ConnPool, db.Statement.ConnPool
					db.ConnPool = inboundCommitFailurePool{sqlDB}
					db.Statement.ConnPool = db.ConnPool
					t.Cleanup(func() { db.ConnPool, db.Statement.ConnPool = pool, statementPool })
				case "activation":
					runtime.addErr = errors.New("test activation failed")
				}
				if _, err := s.saveAndApply(db, runtime, "toggle", inboundToggleData(inbound.Id, !enabled), "", "test", ""); err == nil {
					t.Fatal("failed save reported success")
				}
				_, running := runtime.configs[inbound.Tag]
				if storedInboundEnabled(t, db, inbound.Id) != enabled || running != enabled {
					t.Fatal("failure left database and runtime in different states")
				}
				if failure == "audit" && len(runtime.calls) != 0 {
					t.Fatal("audit failure reached runtime")
				}
				var count int64
				if err := db.Model(&model.Changes{}).Where("actor = ?", "test").Count(&count).Error; err != nil || count != 0 {
					t.Fatalf("failed save left an audit entry: %d %v", count, err)
				}
			})
		}
	}
}

func TestInboundEditActivationFailureRestoresOldListener(t *testing.T) {
	s, db, inbound, runtime := inboundEnableFixture(t, true)
	oldConfig := append([]byte(nil), runtime.configs[inbound.Tag]...)
	runtime.addErr = errors.New("test activation failed")
	data := json.RawMessage(fmt.Sprintf(`{"id":%d,"type":"l2tp","tag":"renamed-vpn","enable":true,"psk":"new-test-only"}`, inbound.Id))
	if _, err := s.saveAndApply(db, runtime, "edit", data, "", "test", ""); err == nil {
		t.Fatal("failed edit reported success")
	}
	var stored model.Inbound
	if err := db.First(&stored, inbound.Id).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Tag != inbound.Tag || !reflect.DeepEqual(runtime.configs[inbound.Tag], oldConfig) || len(runtime.configs) != 1 {
		t.Fatal("failed edit did not restore the old listener and database row")
	}
}

func TestInboundNewRejectsInvalidInitialUserIDs(t *testing.T) {
	s, db, inbound, runtime := inboundEnableFixture(t, true)
	data := json.RawMessage(`{"type":"l2tp","tag":"new-vpn","enable":true,"listen_port":1500,"psk":"test-only"}`)
	if _, err := s.saveAndApply(db, runtime, "new", data, "0) OR (1=1", "test", ""); err == nil {
		t.Fatal("invalid initial user IDs were accepted")
	}
	var count int64
	if err := db.Model(&model.Inbound{}).Count(&count).Error; err != nil || count != 1 || len(runtime.calls) != 0 {
		t.Fatal("failed new inbound changed the database or runtime")
	}
	if _, running := runtime.configs[inbound.Tag]; !running {
		t.Fatal("failed new inbound stopped an existing listener")
	}
}

func TestInboundToggleInMaintenanceKeepsCoreStopped(t *testing.T) {
	s, db, inbound, _ := inboundEnableFixture(t, true)
	configService := &ConfigService{}
	if err := configService.SettingService.SetMaintenance(true); err != nil {
		t.Fatal(err)
	}
	for _, enabled := range []bool{false, true} {
		if _, err := configService.Save("inbounds", "toggle", inboundToggleData(inbound.Id, enabled), "", "test", ""); err != nil {
			t.Fatal(err)
		}
		if storedInboundEnabled(t, db, inbound.Id) != enabled || corePtr.GetInstance() != nil {
			t.Fatal("maintenance toggle started the core or lost state")
		}
	}
	// Even while stopped the list must report the new state.
	list, err := s.GetAll()
	if err != nil || (*list)[0]["enable"] != true {
		t.Fatalf("maintenance list: %v %v", list, err)
	}
}
