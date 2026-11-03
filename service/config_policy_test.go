//go:build !race

// These integration tests start the native NetworkManager, which has the
// upstream race documented by core/race_on_test.go. Policy reference/session
// concurrency is tested separately under -race without starting that manager.
package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"gorm.io/gorm"
)

const policyBaseForTest = `{"log":{"level":"error"},"dns":{"servers":[{"type":"hosts","tag":"hosts","predefined":{"policy.test":["192.0.2.1"]}}]},"route":{"final":"a"}}`

func runningPolicyService(t *testing.T) *ConfigService {
	t.Helper()
	s := lifecycleService(t)
	if err := s.SettingService.SetConfig(policyBaseForTest); err != nil {
		t.Fatal(err)
	}
	if err := database.GetDB().Create(&model.Outbound{Type: "direct", Tag: "a", Options: json.RawMessage(`{}`)}).Error; err != nil {
		t.Fatal(err)
	}
	if err := s.StartCore(); err != nil {
		t.Fatal(err)
	}
	if !corePtr.IsRunning() {
		t.Fatal("test core not running")
	}
	t.Cleanup(func() { _ = s.StopCore() })
	return s
}

func TestPolicySaveFailurePreservesDatabaseAndRuntime(t *testing.T) {
	s := runningPolicyService(t)
	box := corePtr.GetInstance()
	bad := strings.Replace(policyBaseForTest, `"final":"a"`, `"final":"missing"`, 1)
	if _, err := s.Save("config", "edit", []byte(bad), "", "test", ""); err == nil {
		t.Fatal("invalid policy save succeeded")
	}
	stored, err := s.SettingService.GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	if stored != policyBaseForTest || corePtr.GetInstance() != box || !corePtr.IsRunning() {
		t.Fatal("failed preparation replaced stored or running config")
	}
	if _, err := s.Save("outbounds", "del", []byte(`"a"`), "", "test", ""); err == nil {
		t.Fatal("removed required default outbound")
	}
	var count int64
	if err := database.GetDB().Model(&model.Outbound{}).Where("tag = ?", "a").Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("outbound rollback: %d %v", count, err)
	}
}

func TestPolicySaveStorageFailureDoesNotPublishPreparedPolicy(t *testing.T) {
	s := runningPolicyService(t)
	box := corePtr.GetInstance()
	db := database.GetDB()
	const callback = "test:fail_policy_audit"
	if err := db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Table == "changes" {
			tx.AddError(errors.New("test audit write failed"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Create().Remove(callback) })
	updated := strings.ReplaceAll(policyBaseForTest, "192.0.2.1", "192.0.2.2")
	if _, err := s.Save("config", "edit", []byte(updated), "", "test", ""); err == nil {
		t.Fatal("storage failure reported success")
	}
	stored, err := s.SettingService.GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	if stored != policyBaseForTest || corePtr.GetInstance() != box || !corePtr.IsRunning() {
		t.Fatal("storage failure replaced active configuration")
	}
}

func TestPolicySavePublishesBeforeReturning(t *testing.T) {
	s := runningPolicyService(t)
	box := corePtr.GetInstance()
	updated := strings.ReplaceAll(policyBaseForTest, "192.0.2.1", "192.0.2.2")
	if _, err := s.Save("config", "edit", []byte(updated), "", "test", ""); err != nil {
		t.Fatal(err)
	}
	if corePtr.GetInstance() != box {
		t.Fatal("save restarted access listeners")
	}
	// An identical preparation is a no-op only after publication; verify the
	// persisted config can immediately be applied again without restarting.
	if _, err := s.Save("config", "edit", []byte(updated), "", "test", ""); err != nil {
		t.Fatal(err)
	}
}

// Exercise the actual transaction commit boundary, after the candidate and
// audit row have both been prepared. No running policy may be published here.
type policyCommitFailurePool struct{ *sql.DB }

func (p policyCommitFailurePool) BeginTx(ctx context.Context, options *sql.TxOptions) (gorm.ConnPool, error) {
	tx, err := p.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return policyCommitFailureTx{tx}, nil
}

type policyCommitFailureTx struct{ *sql.Tx }

func (tx policyCommitFailureTx) Commit() error {
	_ = tx.Tx.Rollback()
	return errors.New("test policy commit failed")
}

func TestPolicySaveCommitFailurePreservesDatabaseAndRuntime(t *testing.T) {
	s := runningPolicyService(t)
	box := corePtr.GetInstance()
	db := database.GetDB()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	pool, statementPool := db.ConnPool, db.Statement.ConnPool
	db.ConnPool = policyCommitFailurePool{sqlDB}
	db.Statement.ConnPool = db.ConnPool
	t.Cleanup(func() { db.ConnPool, db.Statement.ConnPool = pool, statementPool })
	updated := strings.ReplaceAll(policyBaseForTest, "192.0.2.1", "192.0.2.2")
	if _, err := s.Save("config", "edit", []byte(updated), "", "test", ""); err == nil || !strings.Contains(err.Error(), "commit failed") {
		t.Fatalf("commit failure was not returned: %v", err)
	}
	stored, err := s.SettingService.GetConfig()
	if err != nil || stored != policyBaseForTest || corePtr.GetInstance() != box || !corePtr.IsRunning() {
		t.Fatalf("failed commit changed database or runtime: %q %v", stored, err)
	}
	var count int64
	if err := db.Model(&model.Changes{}).Where("actor = ?", "test").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("failed commit persisted the audit row: %d %v", count, err)
	}
}

func TestPolicySaveWaitsForLifecycleAndApplies(t *testing.T) {
	s := runningPolicyService(t)
	box := corePtr.GetInstance()
	updated := strings.ReplaceAll(policyBaseForTest, "192.0.2.1", "192.0.2.2")
	lifecycleMu.Lock()
	done := make(chan error, 1)
	go func() {
		_, err := s.Save("config", "edit", []byte(updated), "", "test", "")
		done <- err
	}()
	select {
	case err := <-done:
		lifecycleMu.Unlock()
		t.Fatalf("policy save skipped the in-flight lifecycle operation: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	lifecycleMu.Unlock()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("policy save did not finish after the lifecycle operation")
	}
	stored, err := s.SettingService.GetConfig()
	if err != nil || stored != updated || corePtr.GetInstance() != box {
		t.Fatalf("queued policy save was not applied: %q %v", stored, err)
	}
}
