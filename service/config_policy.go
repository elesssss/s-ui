package service

import (
	"encoding/json"
	"time"

	"github.com/alireza0/s-ui/core"
	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
)

// Caller holds lifecycleMu from before reading config until after publication.
// Runtime preparation sees the same transaction that will be committed.
func (s *ConfigService) savePolicy(obj, act string, data json.RawMessage, actor string) ([]string, error) {
	base := string(data)
	stored, err := s.SettingService.GetConfig()
	if err != nil {
		return nil, err
	}
	if obj == "outbounds" {
		base = stored
	}
	tx := database.GetDB().Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.Rollback()
	if obj == "outbounds" {
		err = s.OutboundService.Save(tx, act, data)
	} else {
		err = s.SettingService.SaveConfig(tx, data)
	}
	if err != nil {
		return nil, err
	}
	raw, err := s.getConfigAt(tx, base)
	if err != nil {
		return nil, err
	}
	var update *core.PolicyUpdate
	if corePtr.IsRunning() {
		update, err = corePtr.PreparePolicy(*raw)
		if err != nil {
			return nil, err
		}
		defer update.Abort()
	} else {
		// Even in maintenance, syntax/protocol validation must not be skipped.
		if err = corePtr.ValidateConfig(*raw); err != nil {
			return nil, err
		}
	}
	if err = tx.Create(&model.Changes{DateTime: time.Now().Unix(), Actor: actor, Key: obj, Action: act, Obj: data}).Error; err != nil {
		return nil, err
	}
	if err = tx.Commit().Error; err != nil {
		return nil, err
	}
	LastUpdate.Store(time.Now().Unix())
	if update != nil {
		if err = update.Commit(); err != nil {
			return nil, err
		}
	}
	if update == nil && !s.inMaintenance() {
		if err = s.startCoreLocked(true); err != nil {
			return nil, err
		}
	}
	return []string{obj}, nil
}
